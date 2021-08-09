package commands

const testServerId = "keyring-jfrog-cli-plugin-server-id"

func getTestConf() storeConfiguration {
	return storeConfiguration{
		ServerId: testServerId,
		Url:      "keyring-jfrog-cli-plugin-url",
		User:     "keyring-jfrog-cli-plugin-user",
		Password: "keyring-jfrog-cli-plugin-password",
	}
}

// Compares two string lists. Returns true if they are identical.
func equals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
