109851  11/09/22 16:55:56 operator-sdk init --domain org --repo github.com/movewp3/keycloakclient-controller

109852  11/09/22 16:56:08 operator-sdk create api --group=keycloak --version=v1alpha1 --kind=Keycloak --resource --controller

109857  11/09/22 16:58:45 operator-sdk create api --group=keycloak --version=v1alpha1 --kind=KeycloakRealm --resource --controller

109858  11/09/22 16:58:52 operator-sdk create api --group=keycloak --version=v1alpha1 --kind=KeycloakClient --resource --controller

https://sdk.operatorframework.io/docs/building-operators/golang/migration/


Die Recociler rüberkopieren und den Inhalt aus den Controllern auch rüberkopieren.
Und dann man schauen was noch geht.
Im Migration Dok steht noch einiges, was gemacht werden soll.



make manifests


109554  11/09/22 12:13:47 make client/gen
109555  11/09/22 12:14:18 make code/compile
109622  11/09/22 12:40:27 make setup/mod
109684  11/09/22 13:15:28 make code/compile
109824  11/09/22 16:29:29 make code/compile
109826  11/09/22 16:30:58 make code/compile
109856  11/09/22 16:47:44 make code/gen
109864  11/09/22 16:52:54 make bundle
109873  11/09/22 16:56:18 make bundle
110452  13/09/22 22:13:11 make manifest
110453  13/09/22 22:13:13 make manifests
110624  14/09/22 17:36:48 make manifests 
110626  14/09/22 17:37:07 make generate
112460  16/09/22 11:56:57 make generate
112461  16/09/22 11:57:01 make manifests 
112572  18/09/22 13:35:18 make build
112576  18/09/22 13:39:01 make docker-build
112579  18/09/22 13:42:07 make docker-build
112583  18/09/22 13:42:47 make docker-build
112587  18/09/22 13:55:50 make docker-build
112590  18/09/22 13:56:45 make docker-build
112592  18/09/22 13:59:19 make docker-push
112599  18/09/22 14:03:56 make docker-push

