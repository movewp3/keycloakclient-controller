package e2e

import "time"

const (
	testKeycloakCRName                   = "keycloakx-test"
	testKeycloakRealmCRName              = "keycloakx-realm-test"
	testKeycloakClientCRName             = "keycloakx-client-test"
	testKeycloakConfidentialClientCRName = "a"
	testAuthZKeycloakClientCRName        = "authz-keycloak-client-test"
	testSecondKeycloakClientCRName       = "second-keycloak-client-test"
	pollRetryInterval                    = time.Second * 10
	pollTimeout                          = time.Minute * 1
	keycloakNamespace                    = "keycloak"
)
