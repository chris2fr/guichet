module guichet

go 1.19

require (
	github.com/go-ldap/ldap/v3 v3.4.5
	github.com/go-openapi/runtime v0.26.2
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	github.com/jsimonetti/pwscheme v0.0.0-20220922140336-67a4d090f150
	github.com/minio/minio-go/v7 v7.0.63
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	goauthentik.io/api/v3 v3.2023104.2
	golang.org/x/crypto v0.12.0
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.4 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.4 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/spec v0.20.11 // indirect
	github.com/go-openapi/strfmt v0.21.8 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-openapi/validate v0.22.3 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.mongodb.org/mongo-driver v1.13.1 // indirect
	go.opentelemetry.io/otel v1.17.0 // indirect
	go.opentelemetry.io/otel/metric v1.17.0 // indirect
	go.opentelemetry.io/otel/trace v1.17.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210218202405-ba52d332ba99 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace guichet/views => ./views

replace guichet/utils => ./utils

replace guichet/models => ./models

replace guichet/controllers => ./utils/controllers
