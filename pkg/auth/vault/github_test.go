package vault_test

// Need to find a way to mock GitHub Auth within Vault
// func TestGithubLogin(t *testing.T) {
// 	cluster, role, secret := helpers.CreateTestVault(t)
// 	defer cluster.Cleanup()
//
// 	github := auth.Github{
// 		AccessToken: "test",
// 	}
//
// 	err := github.Authenticate(cluster.Cores[0].Client)
// 	if err != nil {
// 		t.Fatalf("expected no errors but got: %s", err)
// 	}
// }
