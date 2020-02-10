À écrire.

Exemple de config.json pour Deuxfleurs:

```
{
  "http_bind_addr": ":9991",
  "ldap_server_addr": "ldap://bottin2.service.2.cluster.deuxfleurs.fr:389",

  "base_dn": "dc=deuxfleurs,dc=fr",
  "user_base_dn": "ou=users,dc=deuxfleurs,dc=fr",
  "user_name_attr": "cn",
  "group_base_dn": "ou=groups,dc=deuxfleurs,dc=fr",
  "group_name_attr": "cn",

  "admin_account": "cn=admin,dc=deuxfleurs,dc=fr",
  "group_can_admin": "cn=admin,ou=groups,dc=deuxfleurs,dc=fr",
  "group_can_invite": "cn=asso_deuxfleurs,ou=groups,dc=deuxfleurs,dc=fr"
}
```

```
docker run --net host -v $PWD/config.json:/config.json -i lxpz/guichet_amd64:latest 
```
