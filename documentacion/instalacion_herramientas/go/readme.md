

# Instalación de Go (Golang)

Este documento describe cómo instalar Go (Golang) en los sistemas operativos más comunes y cómo verificar que la instalación se haya realizado correctamente.

---

## Requisitos

* Acceso a internet
* Permisos de administrador (sudo o administrador del sistema)

---

## Instalación en Linux

### 1. Descargar Go

Visita el sitio oficial de Go y descarga la versión más reciente para Linux:

[https://go.dev/dl/](https://go.dev/dl/)

O usando `wget` (ejemplo para arquitectura amd64):

```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
```

### 2. Eliminar versiones anteriores (opcional)

```bash
sudo rm -rf /usr/local/go
```

### 3. Extraer Go en `/usr/local`

```bash
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
```

### 4. Configurar variables de entorno

Edita tu archivo `~/.bashrc`, `~/.zshrc` o `~/.profile` y agrega:

```bash
export PATH=$PATH:/usr/local/go/bin
```

Aplica los cambios:

```bash
source ~/.bashrc
```


---

## Verificar la instalación

Ejecuta el siguiente comando en la terminal:

```bash
go version
```

Salida esperada (ejemplo):

```text
go version go1.22.0 linux/amd64
```

---

## Configurar el Workspace (opcional)

Go utiliza módulos, por lo que no es obligatorio definir un `GOPATH`, pero si deseas hacerlo:

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

---

## Crear un proyecto de prueba

```bash
mkdir hola-go
cd hola-go
go mod init hola-go
```

Crea el archivo `main.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hola Go")
}
```

Ejecuta el programa:

```bash
go run main.go
```

---

## Recursos oficiales

* Documentación: [https://go.dev/doc/](https://go.dev/doc/)
* Descargas: [https://go.dev/dl/](https://go.dev/dl/)
* Tutoriales: [https://go.dev/tour/](https://go.dev/tour/)

