# Manual Técnico
## [Desarrollo de un módulo de kernel en C y un daemon en Go para el monitoreo de procesos y contenedores en Linux ]

## 1. Introducción

Este documento describe el diseño técnico, la arquitectura y el funcionamiento
interno del sistema de monitoreo de procesos y contenedores desarrollado
mediante módulos de kernel en C y un daemon en Go para sistemas Linux.

## 2. Alcance del sistema

El sistema permite recolectar métricas de procesos directamente desde el kernel,
exponerlas mediante el sistema de archivos /proc y analizarlas en un daemon
en espacio de usuario, el cual toma decisiones automatizadas para la
estabilización del sistema.

## 3. Arquitectura general

El sistema se compone de los siguientes elementos:

- Módulos de kernel en C (sensores de bajo nivel)
- Interfaz /proc para comunicación kernel–usuario
- Daemon en Go (procesamiento y toma de decisiones)
- Base de datos SQLite
- Sistema de visualización con Grafana
![Arquitectura del proyecto](https://github.com/Desarrollo-Telar/Sistema-de-Financiamiento-ElTelar/blob/master/Backend/project/static/img/corls/image.png)
4. Requisitos del sistema
5. Estructura del proyecto
6. Descripción de componentes
   - Módulos de kernel
   - Daemon en Go
   - Base de datos
   - Automatización
7. Interfaces del sistema
8. Flujo de funcionamiento
9. Instalación y compilación
10. Ejecución del sistema
11. Configuración y umbrales
12. Manejo de errores y depuración
13. Seguridad y consideraciones
14. Limitaciones conocidas
15. Conclusiones
