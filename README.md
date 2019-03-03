# gorncs-api 

API pour le Registre National du Commerce et des Sociétés

- permet la publication des bilans et comptes de résultat au format JSON
- permet l'importation du dépot de fichiers RNCS de l'INPI

## Dépendances

gorncs-api dépend de:

- [Gin Web Framework](http://github.com/gin-gonic/gin)
- [gorncs](http://github.com/chrnin/gorncs)
- [go-sqlite3](http://github.com/mattn/go-sqlite3)
- [go-curl](github.com/andelf/go-curl)
- [libcurl](https://curl.haxx.se/libcurl/)

## Installer gorncs-api
- Installer libcurl4-openssl-dev
- `go get github.com/signaux-faibles/gorncs-api`

## Utilisation 
Lancé sans argument, gorncs-api ouvre un point d'appel sur 127.0.0.1:3000.  

Afin de peupler la base de données, il faut cloner le dépot [RNCS de l'INPI](https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs).

Si le même chemin est configuré pour l'import, les doublons ne seront pas intégrés, il est donc possible de lancer plusieurs fois la même arborescence.

Avant de démarrer l'import, il faut initialiser le schéma de la base avec -initdb.

```
$ gorncs-api -help
Usage of ./gorncs-api:
  -DB string
    	chemin de la base sqlite3 (default "./bilan.db")
  -bind string
    	port d'écoute de l'api (default "127.0.0.1:3000")
  -download
    	synchroniser le dépôt RNCS dans (voir -path, -user et -password)
  -initdb
    	créer une nouvelle base sqlite
  -limit int
    	limiter l'import à n bilans
  -password string
    	mot de passe FTPS RNCS/inpi
  -path string
    	chemin où sont stockés les fichiers RNCS (default ".")
  -scan
    	importer les fichiers
  -siren string
    	restreint l'importation au siren
  -user string
    	utilisateur FTPS RNCS/inpi
  -verbose
    	afficher les informations d'importation

```

### Exemples
#### Initialiser la base de données
WIP

## Appel de l'api
```
$ http :3000/bilan/012345678
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Wed, 27 Feb 2019 14:49:07 GMT
Transfer-Encoding: chunked

[
    {
        "actif_autres_creances_brut": xxxxx,
        "actif_autres_creances_net": xxxxx,
        ...
```