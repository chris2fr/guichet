À écrire.

Exemple de config.json pour Deuxfleurs:

```
{
  "http_bind_addr": ":9991",
  "session_key": "V1BAbmn9VW/wL0EZ6Q8xwhkVq/QVwmwPOtliUlfc0iI=",
  "ldap_server_addr": "ldap://127.0.0.1:389",

  "base_dn": "dc=deuxfleurs,dc=fr",
  "user_base_dn": "ou=users,dc=deuxfleurs,dc=fr",
  "user_name_attr": "cn",
  "group_base_dn": "ou=groups,dc=deuxfleurs,dc=fr",
  "group_name_attr": "cn",

  "group_can_admin": "cn=admin,ou=groups,dc=deuxfleurs,dc=fr",
  "group_can_invite": "cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr"
}
```
