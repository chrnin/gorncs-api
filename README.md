# gorncs-api 

API pour le Registre National du Commerce et des Sociétés

- permet la publication des bilans et comptes de résultat au format JSON
- permet l'importation du dépot de fichiers RNCS de l'INPI

## Dépendances

gorncs-api dépend de:

- [Gin Web Framework](http://github.com/gin-gonic/gin)
- [gorncs](http://github.com/chrnin/gorncs)
- [go-sqlite3](http://github.com/mattn/go-sqlite3)

## Installer gorncs-api
`go get github.com/signaux-faibles/gorncs-api`

## Utilisation 
Lancé sans argument, gorncs-api ouvre un point d'appel sur 127.0.0.1:3000.  
Pour l'utiliser: `http :3000/012345678` vous fournira les bilans du siren 012345678 contenus dans la base sqlite.

Afin de peupler la base de données, il faut cloner le dépot [RNCS de l'INPI](https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs).

Si le même chemin est configuré pour l'import, les doublons ne seront pas intégrés, il est donc possible de lancer plusieurs fois la même arborescence.

```
Usage of ./gorncs-api:
  -DB string
    	chemin de la base sqlite3 (default "./bilan.db")
  -bind string
    	port d'écoute de l'api (default "127.0.0.1:3000")
  -initdb
    	créer une nouvelle base sqlite
  -limit int
    	limiter l'import à n bilans
  -path string
    	chemin où sont stockés les fichiers RNCS (default ".")
  -scan
    	importer les fichiers
  -siren string
    	restreint l'importation au siren
  -verbose
    	afficher les informations d'importation

```
