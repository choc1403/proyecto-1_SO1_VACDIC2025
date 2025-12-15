# Desarrollo de un módulo de kernel en C y un daemon en Go para el monitoreo de procesos y contenedores en Linux

## Descripción general

Este proyecto consiste en el diseño e implementación de un sistema integral de monitoreo, análisis y gestión automatizada de procesos y contenedores en sistemas Linux. La solución combina programación a bajo nivel mediante módulos de kernel escritos en **C**, con programación a alto nivel mediante un **daemon desarrollado en Go**, permitiendo obtener métricas avanzadas directamente desde el kernel y tomar decisiones autónomas para la estabilización del sistema.

El sistema expone información detallada de procesos a través del pseudo-sistema de archivos **/proc**, la cual es consumida y procesada por el daemon en Go. Dicho daemon analiza métricas como uso de CPU, memoria y E/S, persiste información relevante en **SQLite** y la integra con **Grafana** para su visualización mediante dashboards interactivos.

El proyecto está orientado a entornos contenerizados, donde la supervisión proactiva y la gestión automática de recursos resulta crítica para garantizar la estabilidad y el rendimiento del sistema.

---

## Competencias a desarrollar

Al finalizar este proyecto, el estudiante será competente en:

* Diseñar, implementar, compilar y depurar módulos de kernel básicos en C para Linux.
* Comprender y utilizar estructuras de datos fundamentales del kernel relacionadas con la gestión de procesos.
* Crear y gestionar interfaces de comunicación entre el espacio de kernel y el espacio de usuario mediante /proc.
* Consumir interfaces del kernel desde aplicaciones de usuario.
* Desarrollar aplicaciones en Go para interactuar con datos del sistema y parsear archivos.
* Integrar métricas del sistema con herramientas de visualización como Grafana.
* Aplicar buenas prácticas de programación tanto en el entorno del kernel como en aplicaciones de usuario.

---

## Objetivos

### Objetivo general

Desarrollar un sistema integrado que consta de módulos de kernel en C para Linux, capaces de listar procesos activos (generales y asociados a contenedores) y exponer su información vía /proc, junto con una aplicación daemon en Go que interprete, procese y presente estos datos de forma estructurada y amigable, tomando decisiones automáticas para la estabilización del sistema.

### Objetivos específicos

Al finalizar el proyecto, se deberá ser capaz de:

1. Desarrollar e implementar módulos de kernel funcionales:

   * Crear módulos en C que puedan cargarse y descargarse dinámicamente sin causar inestabilidad en el sistema.

2. Extraer información de procesos desde el kernel:

   * Acceder y recorrer estructuras internas del kernel para obtener datos como PID, nombre del proceso, uso de CPU, memoria y E/S.

3. Crear una interfaz de kernel en /proc:

   * Implementar archivos virtuales en /proc que muestren la información recolectada de los procesos.

4. Desarrollar un daemon en Go:

   * Implementar una aplicación que lea y parsee los archivos /proc generados por los módulos del kernel.

5. Estabilizar el sistema mediante análisis automatizado:

   * Desarrollar un sistema de monitoreo que analice métricas en tiempo real y ejecute acciones correctivas automáticas (detención y eliminación selectiva de contenedores) cuando se superen umbrales predefinidos.

---

## Enunciado del proyecto

El objetivo principal es diseñar e implementar un sistema integral para la monitorización proactiva, el análisis automatizado y la gestión inteligente de contenedores en entornos Linux, combinando el acceso directo a las estructuras del kernel con lógica de alto nivel para la toma de decisiones autónomas.

---

## Problema a resolver

En la administración de sistemas Linux y el desarrollo de aplicaciones contenerizadas, obtener información detallada sobre los procesos en ejecución y actuar de forma proactiva representa un desafío importante. Herramientas tradicionales como `ps` o `docker stats` proporcionan información limitada y no permiten acceder directamente a las estructuras internas del kernel ni automatizar acciones correctivas avanzadas.

Este proyecto propone una solución integral que:

* Expone métricas avanzadas de procesos y contenedores directamente desde el kernel.
* Centraliza y analiza dichas métricas en un daemon de usuario.
* Automatiza la gestión de contenedores en función del consumo de recursos.
* Almacena información histórica para análisis posterior.
* Presenta los datos de forma visual mediante dashboards en Grafana.

Para validar el funcionamiento del sistema, se implementan **cronjobs** que generan contenedores de prueba cada minuto, simulando cargas variables y permitiendo evaluar la efectividad de las acciones correctivas.

---

## Arquitectura del sistema

El sistema está compuesto por los siguientes elementos:

### 1. Módulo de kernel para procesos de contenedores (C)

* Actúa como sensor de bajo nivel.
* Accede directamente a las estructuras internas del kernel.
* Captura métricas detalladas de procesos asociados a contenedores:

  * CPU
  * Memoria
  * E/S
* Expone la información mediante archivos virtuales en /proc.

### 2. Módulo de kernel para procesos generales (C)

* Similar al módulo anterior, pero enfocado en todos los procesos del sistema.
* Permite obtener una visión global del consumo de recursos del sistema.

### 3. Daemon en Go

Funciona como el núcleo lógico del sistema y se encarga de:

* Leer y parsear los datos expuestos en /proc por los módulos del kernel.
* Analizar métricas en tiempo real.
* Tomar decisiones autónomas basadas en umbrales y patrones definidos:

  * Detener contenedores.
  * Eliminar contenedores que excedan límites de recursos.
* Ejecutar scripts de automatización durante la ejecución.
* Persistir datos relevantes en una base de datos SQLite.

### 4. Cronjob de generación de carga

* Ejecuta scripts que crean contenedores Docker cada minuto.
* Simula escenarios de carga continua y variable.
* Permite validar el comportamiento y la estabilidad del sistema.

### 5. Dashboard en Grafana

* Consume los datos almacenados por el daemon en Go.
* Presenta métricas del sistema de forma visual e interactiva.
* Facilita el análisis del rendimiento y la toma de decisiones.

---

## Tecnologías utilizadas

* Lenguaje C (desarrollo de módulos de kernel)
* Go (desarrollo del daemon)
* Linux Kernel Modules
* /proc filesystem
* Docker
* SQLite
* Grafana
* Cron

---

## Alcance del proyecto

Este proyecto integra conceptos fundamentales de sistemas operativos, programación a bajo nivel y desarrollo de servicios de usuario, proporcionando una solución realista a un problema común en entornos contenerizados: la monitorización y estabilización autónoma del sistema.

Además de su valor práctico, el proyecto fortalece la comprensión del funcionamiento interno del kernel de Linux y la interacción entre el espacio de kernel y el espacio de usuario.
