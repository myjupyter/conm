package registry

// import (
// 	"github.com/myjupyter/conm/internal/config"
// 	"github.com/myjupyter/conm/internal/secret"
// )
//
// func NewKeyringSecretRegistry(refs ...secret.Reference) (*SecretRegistry[config.Secret], error) {
// 	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPath())
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	r := &SecretRegistry[config.Secret]{file: file}
// 	for _, ref := range refs {
// 		r.refs.link(ref.SecretRef())
// 	}
//
// 	return r, nil
// }
//
