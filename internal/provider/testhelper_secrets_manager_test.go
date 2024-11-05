package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	jwtSigningSecret = "secret-which-does-not-matter"
)

func NewTestSecretsManager() *testSecretsManager {
	return &testSecretsManager{
		clientSideInformation: ClientSideInformation{
			orgEncryptionKeys: map[string]symmetrickey.Key{},
		},
		issuedJwtTokens:    map[string]string{},
		knownClients:       map[string]Clients{},
		knownOrganizations: map[string]struct{}{},
		projectsStore:      map[string]models.Project{},
		secretsStore:       map[string]webapi.Secret{},
	}
}

type testSecretsManager struct {
	clientSideInformation ClientSideInformation
	issuedJwtTokens       map[string]string
	knownClients          map[string]Clients
	knownOrganizations    map[string]struct{}
	projectsStore         map[string]models.Project
	secretsStore          map[string]webapi.Secret
}

type Clients struct {
	ClientID         string
	ClientSecret     string
	EncryptedPayload string
	OrganizationID   string
}

type ClientSideInformation struct {
	orgEncryptionKeys map[string]symmetrickey.Key
}

type CreateAccessTokenRequest struct {
	EncryptedPayload string
}

type CreateAccessTokenResponse struct {
	ClientSecret string
	Id           string
	Object       string
}

func (tsm *testSecretsManager) Run(ctx context.Context, serverPort int) {
	handler := mux.NewRouter()
	handler.HandleFunc("/api/organizations/{orgId}/secrets", tsm.handlerCreateGetSecret).Methods("POST", "GET")
	handler.HandleFunc("/api/organizations/{orgId}/projects", tsm.handlerCreateProject).Methods("POST")
	handler.HandleFunc("/api/projects/{projectId}", tsm.handlerGetProject).Methods("GET")
	handler.HandleFunc("/api/projects/{projectId}", tsm.handlerEditProject).Methods("PUT")
	handler.HandleFunc("/api/projects/delete", tsm.handlerDeleteProject).Methods("POST")
	handler.HandleFunc("/api/secrets/{secretId}", tsm.handlerGetSecret).Methods("GET")
	handler.HandleFunc("/api/secrets/{secretId}", tsm.handlerEditSecret).Methods("PUT")
	handler.HandleFunc("/api/secrets/delete", tsm.handlerDeleteSecret).Methods("POST")
	handler.HandleFunc("/identity/connect/token", tsm.handlerLogin).Methods("POST")

	server := &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf(":%d", serverPort),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %v\n", err)
		}
	}()

	<-ctx.Done()

	server.Shutdown(context.Background())
}

func (tsm *testSecretsManager) ClientCreateNewOrganization() (string, error) {
	encryptionKey, err := generateOrganizationKey()
	if err != nil {
		return "", err
	}

	orgId := uuid.New().String()
	tsm.knownOrganizations[orgId] = struct{}{}

	tsm.clientSideInformation.orgEncryptionKeys[orgId] = *encryptionKey
	return orgId, nil
}

func (tsm *testSecretsManager) ClientCreateAccessToken(orgId string) (string, error) {
	orgKey, v := tsm.clientSideInformation.orgEncryptionKeys[orgId]
	if !v {
		return "", fmt.Errorf("organization not found")
	}

	accessTokenEncryptionKey, err := generateAccessTokenEncryptionKey()
	if err != nil {
		return "", err
	}

	encryptedPayload, err := encryptPayload(accessTokenEncryptionKey, orgKey)
	if err != nil {
		return "", fmt.Errorf("error encrypting payload: %w", err)
	}

	request := CreateAccessTokenRequest{
		EncryptedPayload: encryptedPayload,
	}

	response, err := tsm.createAccessToken(orgId, request)
	if err != nil {
		return "", fmt.Errorf("error creating access token: %w", err)
	}

	return fmt.Sprintf("0.%s.%s:%s", response.Id, response.ClientSecret, base64.StdEncoding.EncodeToString(accessTokenEncryptionKey)), nil
}

func (tsm *testSecretsManager) createAccessToken(orgId string, request CreateAccessTokenRequest) (*CreateAccessTokenResponse, error) {
	clientSecretBytes, err := generateClientSecret()
	if err != nil {
		return nil, err
	}

	client := Clients{
		ClientID:         uuid.New().String(),
		ClientSecret:     base64.StdEncoding.EncodeToString(clientSecretBytes),
		OrganizationID:   orgId,
		EncryptedPayload: request.EncryptedPayload,
	}
	tsm.knownClients[client.ClientID] = client

	return &CreateAccessTokenResponse{
		Id:           client.ClientID,
		ClientSecret: client.ClientSecret,
	}, nil
}

func (tsm *testSecretsManager) handlerLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	providedClientId := r.FormValue("client_id")
	client, clientExists := tsm.knownClients[providedClientId]
	if !clientExists {
		http.Error(w, "Invalid client id", http.StatusBadRequest)
		return
	}

	providedClientSecret := r.FormValue("client_secret")
	if client.ClientSecret != providedClientSecret {
		http.Error(w, "Invalid client secret", http.StatusBadRequest)
		return
	}

	_, orgExists := tsm.knownOrganizations[client.OrganizationID]
	if !orgExists {
		http.Error(w, "Invalid organization", http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &embedded.MachineAccountClaims{
		Organization: client.OrganizationID,
	})
	jwtAccessToken, err := token.SignedString([]byte(jwtSigningSecret))
	if err != nil {
		http.Error(w, "error generating jwt token: %w", http.StatusBadRequest)
		return
	}

	tsm.issuedJwtTokens[jwtAccessToken] = client.ClientID

	response := webapi.MachineTokenResponse{
		AccessToken:      jwtAccessToken,
		ExpireIn:         3600,
		TokenType:        "Bearer",
		EncryptedPayload: string(client.EncryptedPayload),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerCreateGetSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tsm.handlerGetSecrets(w, r)
	} else {
		tsm.handlerCreateSecret(w, r)
	}
}

func (tsm *testSecretsManager) handlerCreateProject(w http.ResponseWriter, r *http.Request) {
	orgId := mux.Vars(r)["orgId"]

	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var projectCreationRequest webapi.CreateProjectRequest
	if err := json.Unmarshal(body, &projectCreationRequest); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	project := models.Project{
		ID:             uuid.New().String(),
		Name:           projectCreationRequest.Name,
		OrganizationID: orgId,
		CreationDate:   time.Now(),
		RevisionDate:   time.Now(),
		Object:         string(models.ObjectProject),
	}
	tsm.projectsStore[project.ID] = project

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerCreateSecret(w http.ResponseWriter, r *http.Request) {
	orgId := mux.Vars(r)["orgId"]
	_, v := tsm.knownOrganizations[orgId]
	if !v {
		http.Error(w, "Invalid organization", http.StatusBadRequest)
		return
	}

	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var secretCreationRequest webapi.CreateSecretRequest
	if err := json.Unmarshal(body, &secretCreationRequest); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	projects := []models.Project{}
	for _, v := range secretCreationRequest.ProjectIDs {
		project, projectExists := tsm.projectsStore[v]
		if !projectExists {
			http.Error(w, "Project not found", http.StatusBadRequest)
			return
		}
		projects = append(projects, project)
	}

	secret := webapi.Secret{
		SecretSummary: webapi.SecretSummary{
			ID:             uuid.New().String(),
			OrganizationID: orgId,
			Key:            secretCreationRequest.Key,
			CreationDate:   time.Now(),
			RevisionDate:   time.Now(),
			Projects:       projects,
			Read:           true,
			Write:          true,
		},
		Value:  secretCreationRequest.Value,
		Note:   secretCreationRequest.Note,
		Object: string(models.ObjectSecret),
	}
	tsm.secretsStore[secret.ID] = secret

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerGetSecrets(w http.ResponseWriter, r *http.Request) {
	orgId := mux.Vars(r)["orgId"]

	secretList := webapi.SecretsWithProjectsList{}
	for _, v := range tsm.secretsStore {
		if v.OrganizationID != orgId {
			continue
		}

		sum := webapi.SecretSummary{
			ID:             v.ID,
			OrganizationID: v.OrganizationID,
			Key:            v.Key,
			CreationDate:   v.CreationDate,
			RevisionDate:   v.RevisionDate,
			Projects:       v.Projects,
			Read:           v.Read,
			Write:          v.Write,
		}
		secretList.Secrets = append(secretList.Secrets, sum)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secretList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerGetProject(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	projectId := mux.Vars(r)["projectId"]
	project, projectExists := tsm.projectsStore[projectId]
	if !projectExists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerGetSecret(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	secretId := mux.Vars(r)["secretId"]
	secret, secretExists := tsm.secretsStore[secretId]
	if !secretExists {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerEditProject(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	projectId := mux.Vars(r)["projectId"]
	project, projectExists := tsm.projectsStore[projectId]
	if !projectExists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var projectCreationRequest webapi.CreateProjectRequest
	if err := json.Unmarshal(body, &projectCreationRequest); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	project.RevisionDate = time.Now()
	project.Name = projectCreationRequest.Name
	tsm.projectsStore[projectId] = project

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerEditSecret(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	secretId := mux.Vars(r)["secretId"]
	secret, secretExists := tsm.secretsStore[secretId]
	if !secretExists {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var secretCreationRequest webapi.CreateSecretRequest
	if err := json.Unmarshal(body, &secretCreationRequest); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	secret.RevisionDate = time.Now()
	secret.Key = secretCreationRequest.Key
	secret.Value = secretCreationRequest.Value
	secret.Note = secretCreationRequest.Note
	tsm.secretsStore[secretId] = secret

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (tsm *testSecretsManager) handlerDeleteProject(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var IDs []string
	if err := json.Unmarshal(body, &IDs); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	for _, v := range IDs {
		delete(tsm.projectsStore, v)
	}
	w.WriteHeader(http.StatusOK)
}

func (tsm *testSecretsManager) handlerDeleteSecret(w http.ResponseWriter, r *http.Request) {
	err := tsm.checkAuthentication(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var IDs []string
	if err := json.Unmarshal(body, &IDs); err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	for _, v := range IDs {
		delete(tsm.secretsStore, v)
	}
	w.WriteHeader(http.StatusOK)
}

func (tsm *testSecretsManager) checkAuthentication(authorization string) error {
	if authorization == "" {
		return fmt.Errorf("missing Authorization header")
	}

	authorization = authorization[7:]
	clientID, jwtTokenKnown := tsm.issuedJwtTokens[authorization]
	if !jwtTokenKnown {
		return fmt.Errorf("invalid token")
	}

	_, clientKnown := tsm.knownClients[clientID]
	if !clientKnown {
		return fmt.Errorf("client doesn't exist anymore")
	}
	return nil
}

func encryptPayload(accessTokenEncryptionKey []byte, orgEncryptionKey symmetrickey.Key) (string, error) {
	payload := webapi.MachineTokenEncryptedPayload{
		EncryptionKey: base64.StdEncoding.EncodeToString(orgEncryptionKey.Key),
	}
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	accessKeyEncryptionKey, err := keybuilder.DeriveFromAccessTokenEncryptionKey(accessTokenEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("error deriving access key: %w", err)
	}

	encryptedPayloadRaw, err := crypto.Encrypt(payloadRaw, *accessKeyEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("error encrypting payload: %w", err)
	}
	return encryptedPayloadRaw.String(), nil
}

func generateOrganizationKey() (*symmetrickey.Key, error) {
	encryptionKey := make([]byte, 64)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error generating organization key: %w", err)
	}
	return symmetrickey.NewFromRawBytes(encryptionKey)
}

func generateAccessTokenEncryptionKey() ([]byte, error) {
	encryptionKey := make([]byte, 64)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error generating organization key: %w", err)
	}
	return encryptionKey, nil
}

func generateClientSecret() ([]byte, error) {
	clientSecretBytes := make([]byte, 64)
	_, err := rand.Read(clientSecretBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating client secret: %w", err)
	}
	return clientSecretBytes, nil
}
