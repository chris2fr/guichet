# Intégration de Guichet dans un environnement de dev/test

## Dev process

On utilise `docker compose` pour mettre en place l'infrastructure dont dépend Guichet, que l'on développe. (On rajoutera Garage dedans plus tard.)

On ne met pas Guichet dans le `compose` pour pouvoir itérer plus rapidement : un `go build` et on a la nouvelle version, sans avoir restart les dépendances (Bottin, Consul...).

## Notes

* Bien récupérer le password `admin` dans les logs de 1er lancement de Bottin : il ne sera pas réaffiché.
* Identifiant de l'admin sur Guichet : `cn=admin,dc=bottin,dc=eu` because il n'est pas dans `ou=users,dc=bottin,dc=eu` qui est l'organisation par défaut dans laquelle on va chercher les utilisateurs.

## TODO 

* Bridger Garage/S3 (pour le moment ne sert que pour les avatars dans l'annuaire)