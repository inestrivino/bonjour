# bonjour

- /build
  - /ci
  - /package
- /cmd
  - main.go
- /docs
- /examples
- /internal
  - /config
    - config.go
    - config_test.go
  - modules
    - /events
    - /quotes
    - /rss
    - /weather
  - /ui
- /scripts
- .gitignore
- go.mod
- go.sum
- LICENSE
- README.md

If you run the script as `NO_COLOR=1 go run cmd/main.go` you will be able to display the result without color or style.


Para que la dashboard cargue instantáneamente, no debería hacer peticiones HTTPS secuenciales, si no usar GoRoutines en paralelo utilizando sync.WaitGroup o canales. Si una API tarda más de 500ms entonces muestra un fallback mientras lo sigue intentando en el backend (con un máximo). 

Para la instalación deberíamos tener tanto un script install.sh (e uninstall.sh por supuesto) que compile el binario y lo mueva a /usr/local/bin/dashboard, como, Para la ejecución diaria, el dashboard puede guardar una marca de tiempo en ~/.config/dashboard/.last_run. Al abrir la terminal, verifica si la fecha de .last_run es diferente a la de hoy. Si es igual, finaliza en silencio; si es distinta, se muestra y actualiza la fecha.

No olvides crear un comando --help (todas las opciones de comando), y un comando --mini (más conciso)
