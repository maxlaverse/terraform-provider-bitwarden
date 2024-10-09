package webapi

import (
	"fmt"
	"runtime"
	"strings"
)

type deviceInfo struct {
	deviceType       string
	deviceName       string
	deviceIdentifier string
	deviceVersion    string
	userAgent        string
}

type deviceInfoWithOfficialFallback struct {
	deviceInfo
	official deviceInfo
}

func DeviceInformation(deviceId, providerVersion string) deviceInfoWithOfficialFallback {
	// Bitwarden has a hard-coded list of device types which can be found at:
	//   => https://github.com/bitwarden/server/blob/main/src/Core/Enums/DeviceType.cs
	//
	// Authenticating using username/password is not documented, and seems to be restricted to
	// the officially recognized devices. To keep a one-to-one feature correspondence between the
	// CLI and the embedded client, we're going to fake our device information when an API call
	// requires it, while showing our real identify as often as possible.
	// Still, we're going to strongly encourage users to rely on the API key authentication, since
	// it's the only officially supported way to authenticate.
	var correspondingOfficialDeviceType, correspondingOfficialDeviceName string
	correspondingOfficialDeviceVersion := "2024.9.0"

	switch os := runtime.GOOS; os {
	case "windows":
		correspondingOfficialDeviceType = "23"
		correspondingOfficialDeviceName = "windows"
	case "darwin":
		correspondingOfficialDeviceType = "24"
		correspondingOfficialDeviceName = "macos"
	case "linux":
		correspondingOfficialDeviceType = "25"
		correspondingOfficialDeviceName = "linux"
	}

	return deviceInfoWithOfficialFallback{
		deviceInfo: deviceInfo{
			deviceType:       "21", // SDK
			deviceName:       "Bitwarden_Terraform_Provider",
			deviceIdentifier: deviceId,
			deviceVersion:    providerVersion,
			userAgent:        fmt.Sprintf("Bitwarden_Terraform_Provider/%s (%s)", providerVersion, strings.ToUpper(runtime.GOOS)),
		},
		official: deviceInfo{
			deviceType:       correspondingOfficialDeviceType,
			deviceName:       correspondingOfficialDeviceName,
			deviceIdentifier: deviceId,
			deviceVersion:    correspondingOfficialDeviceVersion,
			userAgent:        fmt.Sprintf("Bitwarden_CLI/%s (%s)", correspondingOfficialDeviceVersion, strings.ToUpper(correspondingOfficialDeviceName)),
		},
	}
}
