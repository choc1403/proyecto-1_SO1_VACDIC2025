

# Instalación de C en Ubuntu

Este documento explica cómo instalar el compilador de C en Ubuntu, configurar el entorno de desarrollo y verificar que todo funcione correctamente.

---

## Requisitos

* Sistema operativo Ubuntu
* Acceso a internet
* Permisos de administrador (sudo)

---

## Actualizar el sistema

Antes de instalar, se recomienda actualizar la lista de paquetes:

```bash
sudo apt update
sudo apt upgrade -y
```

---

## Instalar el compilador de C (GCC)

El lenguaje C se compila usando **GCC (GNU Compiler Collection)**.

Instala GCC con el siguiente comando:

```bash
sudo apt-get install make

sudo apt install build-essential -y
```

Este paquete incluye:

* `gcc` (compilador de C)
* `g++` (compilador de C++)
* `make`
* Librerías estándar necesarias

---

## Verificar la instalación

Comprueba que el compilador esté instalado correctamente:

```bash
gcc --version
```

Salida esperada (ejemplo):

```text
gcc (Ubuntu 13.2.0) 13.2.0
```

---

## Crear un programa de prueba en C

### 1. Crear el archivo

```bash
nano hola.c
```

Contenido del archivo:

```c
#include <stdio.h>

int main() {
    printf("Hola C en Ubuntu\n");
    return 0;
}
```

Guarda el archivo con `Ctrl + O` y sal con `Ctrl + X`.

---

## Compilar el programa

```bash
gcc hola.c -o hola
```

Esto generará un ejecutable llamado `hola`.

---

## Ejecutar el programa

```bash
./hola
```

Salida esperada:

```text
Hola C en Ubuntu
```

---

## Instalar herramientas adicionales (opcional)

### Debugger (gdb)

```bash
sudo apt install gdb -y
```

### Páginas de manual y documentación

```bash
sudo apt install manpages-dev -y
```

---

## Desinstalar GCC (opcional)

Si deseas eliminar GCC:

```bash
sudo apt remove build-essential -y
```

---

## Recursos útiles

* Documentación de GCC: [https://gcc.gnu.org/](https://gcc.gnu.org/)
* Manual de C: [https://en.cppreference.com/w/c](https://en.cppreference.com/w/c)
* Ubuntu Packages: [https://packages.ubuntu.com/](https://packages.ubuntu.com/)

