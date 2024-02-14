package model

// Constants for a community Keycloak installation
const (
	ApplicationName                  = "keycloak"
	DefaultControllerNamespace       = "keycloak"
	AdminUsernameProperty            = "ADMIN_USERNAME"
	AdminPasswordProperty            = "ADMIN_PASSWORD"
	ClientName                       = "KEYCLOAKCLIENT_CONTROLLER_NAME"
	ClientPassword                   = "KEYCLOAKCLIENT_CONTROLLER_PASSWORD"
	KeycloakClientSecretSeed         = "SECRET_SEED"
	SecretSeedSecretName             = "credential-keycloak-client-secret-seed"
	SALT                             = "803%%1Pas$3cow++#"
	ServingCertSecretName            = "sso-x509-https-secret" // nolint
	ClientSecretName                 = ApplicationName + "-client-secret"
	ClientSecretClientIDProperty     = "CLIENT_ID"
	ClientSecretClientSecretProperty = "CLIENT_SECRET"
)

var PodLabels = map[string]string{}
