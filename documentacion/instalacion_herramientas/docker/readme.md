
# Guía de Instalación de Docker Engine en Ubuntu

Esta guía proporciona los pasos oficiales y recomendados para instalar **Docker Engine, Docker CLI y Containerd** en un sistema Ubuntu, asegurando que obtienes la última versión estable directamente desde el repositorio oficial de Docker.

##  Requisitos Previos

* Un sistema operativo **Ubuntu** (cualquier versión LTS, como 20.04 o 22.04).
* Acceso a la terminal.
* Permisos de `sudo`.

---

## Pasos de Instalación

Sigue los siguientes comandos en tu terminal.

### Paso 1: Actualizar el Sistema e Instalar Paquetes Necesarios

Asegúrate de que tus paquetes locales están actualizados e instala las dependencias necesarias para usar repositorios sobre HTTPS.

```bash
# Actualizar el índice de paquetes local
sudo apt update

# Instalar paquetes para gestionar repositorios de forma segura
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
```

### Paso 2: Añadir la Clave GPG Oficial de Docker

Esto verifica la autenticidad de los paquetes descargados de Docker.

```bash
# Descargar la clave GPG de Docker
curl -fsSL [https://download.docker.com/linux/ubuntu/gpg](https://download.docker.com/linux/ubuntu/gpg) | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
```


### Paso 3: Configurar el Repositorio Estable de Docker

Añade el repositorio oficial de Docker a tus fuentes de APT.

```bash
# Agregar el repositorio de Docker (usando la arquitectura de tu sistema y la distribución de Ubuntu)
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] [https://download.docker.com/linux/ubuntu](https://download.docker.com/linux/ubuntu) \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

### Paso 4: Instalar Docker Engine

Actualiza el índice de paquetes nuevamente (ahora incluirá el repositorio de Docker) e instala los componentes principales.

```bash
# Actualizar el índice de paquetes (para incluir Docker)
sudo apt update

# Instalar Docker Engine, CLI y containerd.io
sudo apt install docker-ce docker-ce-cli containerd.io
```

-----

## Verificación de la Instalación

Comprueba que Docker se haya instalado correctamente y que el servicio esté funcionando.

### 1\. Comprobar el Estado del Servicio

```bash
sudo systemctl status docker
```

> **Resultado esperado:** Deberías ver `Active: active (running)`.

### 2\. Ejecutar la Imagen de Prueba `hello-world`

```bash
sudo docker run hello-world
```

> **Resultado esperado:** Docker descarga y ejecuta una pequeña imagen de prueba, imprimiendo un mensaje de confirmación que indica que la instalación funciona.

-----

## Configuración Post-Instalación (Opcional, pero Recomendado)

Por defecto, solo el usuario `root` y los usuarios en el grupo `docker` pueden ejecutar comandos Docker. Para evitar usar `sudo` antes de cada comando, añade tu usuario al grupo `docker`.

### 1\. Añadir tu Usuario al Grupo Docker

```bash
# Reemplaza $USER con tu nombre de usuario si no se rellena automáticamente
sudo usermod -aG docker $USER
```

### 2\. Aplicar los Cambios

Para que la pertenencia al nuevo grupo surta efecto, debes **cerrar la sesión y volver a iniciarla** o reiniciar tu sistema.

```bash
# Cierra y vuelve a abrir tu sesión de terminal o reinicia el sistema
# Luego, verifica con:
docker run hello-world
```

Si el comando se ejecuta sin `sudo`, la configuración fue exitosa.

