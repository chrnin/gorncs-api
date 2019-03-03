# gorncs-api 

**Dépot obsolète: voir https://github.com/signaux-faibles/gorncs-api**

API pour le Registre National du Commerce et des Sociétés

Permet la consultation des bilans et comptes de résultat au format JSON

## Dépendances

gorncs-api dépend de:

- [Gin Web Framework](http://github.com/gin-gonic/gin)
- [gorncs](http://github.com/chrnin/gorncs)
- [go-sqlite3](http://github.com/mattn/go-sqlite3)

## Installer gorncs-api
`go get github.com/chrnin/gorncs-api`

## Utilisation 
Lancé sans argument, gorncs-api ouvre un point d'appel sur l'interface localhost, port 3000.  
Pour l'utiliser: `http :3000/012345678` vous fournira les bilans du siren 012345678 contenus dans la base sqlite

Afin de peupler la base de données, il faut cloner le dépot [RNCS de l'INPI](https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs).

```
Usage of ./gorncs-api:
  -DB string
        sqlite3 database file
  -bind string
        Listen and serve on (default "127.0.0.1:3000")
  -dial string
        MongoDB dial URL (default "localhost")
  -path string
        RNCS root path (default ".")
  -scanner
        Scan and import everything below the root path, doesn't run API endpoint
```
