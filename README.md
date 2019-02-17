# gorncs-api 

API pour le Registre National du Commerce et des Sociétés

Permet la consultation des bilans et comptes de résultat au format JSON

## Dépendances

gorncs-api dépend de:

- [Gin Web Framework](http://github.com/gin-gonic/gin)
- [gorncs](http://github.com/chrnin/gorncs)
- [mongodb](https://www.mongodb.com/)

## Installer GoRNCS-cli
Installer MongoDB par votre moyen préféré.  

`go get github.com/chrnin/gorncs-api`

## Utilisation
Lancé sans argument, gorncs-api ouvre un point d'appel sur l'interface localhost, port 3000.  
Pour l'utiliser: `http :3000/012345678` vous fournira les bilans contenus dans la base mongodb `inpi` et la collection `bilan`

Afin de peupler la base de données, il faut cloner le dépot [RNCS de l'INPI](https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs).

```
  gorncs-api -help
    Usage of ./gorncs-api:
  -C string
        MongoDB collection (default "bilan")
  -DB string
        MongoDB database (default "inpi")
  -dial string
        MongoDB dial URL (default "localhost")
  -path string
        RNCS root path (default ".")
  -scanner
        Scan and import the root directory
```
// TODO 