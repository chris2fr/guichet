# Intégration de Guichet dans un environnement de dev/test

## Dev process

On utilise `docker compose` pour mettre en place l'infrastructure dont dépend Guichet, que l'on développe. (On rajoutera Garage dedans plus tard.)

On ne met pas Guichet dans le `compose` pour pouvoir itérer plus rapidement : un `go build` et on a la nouvelle version, sans avoir restart les dépendances (Bottin, Consul...).

## Notes

* Bien récupérer le password `admin` dans les logs de 1er lancement de Bottin : il ne sera pas réaffiché.