package main

func getUbuntuReleaseName(version string) string {
	switch version {
	case "20.04":
		return "focal"
	case "22.04":
		return "jammy"
	case "24.04":
		return "noble"
	default:
		return version
	}
}
