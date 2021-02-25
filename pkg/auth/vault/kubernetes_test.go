package vault_test

// Need to find a way to mock k8s Auth within Vault
// func TestGithubLogin(t *testing.T) {
// 	cluster, role, secret := helpers.CreateTestVault(t)
// 	defer cluster.Cleanup()
//
// 	k8s := vault.NewK8sAuth("", "", "")
//
// 	err := k8s.Authenticate(cluster.Cores[0].Client)
// 	if err != nil {
// 		t.Fatalf("expected no errors but got: %s", err)
// 	}
// }
