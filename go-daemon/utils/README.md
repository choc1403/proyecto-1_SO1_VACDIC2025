#  Package `utils`

El paquete **utils** proporciona funciones auxiliares esenciales para el daemon, enfocadas en:

* Limpieza y normalización de datos (especialmente JSON).
* Lectura segura de archivos del sistema (como los de `/proc`).
* Parsing de porcentajes de memoria.
* Ejecución de comandos del sistema *(si deseas, puedo agregar esta sección si tu package lo usa)*.

Estas utilidades permiten que otras partes del sistema (como el módulo de CPU, lógica o monitoreo de contenedores) trabajen con datos limpios, seguros y en un formato consistente.

---

# Funciones principales

##  `var TrailingCommaRe = regexp.MustCompile(",\\s*([\\]\\}])")`

Expresión regular utilizada para detectar **comas finales no válidas en JSON**.

Ejemplo de JSON no estándar:

```json
{
    "key": "value",
}
```

Este tipo de comas no son válidas en JSON estricto, por lo que deben eliminarse antes de hacer `json.Unmarshal`.
La expresión regular identifica casos como:

* `"value", }`
* `"item1", ]`

Y permite sanearlos correctamente.

---

## `func SanitizeJSON(b []byte) []byte`

Limpia y normaliza JSON **malformado** eliminando comas finales inválidas.

### Proceso:

1. Recibe un `[]byte` con contenido JSON.
2. Aplica la expresión regular `TrailingCommaRe`.
3. Reemplaza secuencias como `", }"` → `"}"`.
4. Devuelve un JSON válido que puede parsearse con seguridad mediante `json.Unmarshal`.

### ¿Por qué es útil?

Muchos sistemas generan JSON con trailing commas, lo que rompe el parsing.
Esta función garantiza robustez en el daemon ante este tipo de errores.

---

##  `func ReadProcFile(path string) ([]byte, error)`

Lee archivos del sistema, especialmente los ubicados en:

```
/proc
```

Estos archivos contienen información del kernel, CPU, contenedores montados, etc.

### Proceso:

1. Abre el archivo indicado en `path`.
2. Usa `defer f.Close()` para cerrar el archivo correctamente.
3. Lee su contenido con un límite de:

   ```
   10 MB (10<<20)
   ```

   Esto protege el daemon si un archivo es inesperadamente grande.
4. Retorna un slice de bytes con su contenido o un error.

### Ventajas:

* Seguro ante archivos grandes.
* Ideal para la lectura repetitiva que hace el daemon.
* Funciona de forma uniforme para cualquier archivo del `/proc`.

---

##  `func ParseMemPct(s string) (float64, error)`

Convierte una cadena que representa un porcentaje de memoria a un número `float64`.

### Proceso:

1. Limpia espacios con `strings.TrimSpace`.
2. Si la cadena queda vacía → retorna `0.0`.
3. Convierte usando:

   ```go
   strconv.ParseFloat(s, 64)
   ```

### Uso típico:

En la lógica del daemon, esta función se usa para procesar valores provenientes de archivos como `/proc/cont`, donde la memoria puede venir como:

```
"45.3"
"89"
" 12.7 "
```

La función garantiza que siempre retorne un valor numérico válido.

---

#  Conclusión

El paquete `utils` proporciona funciones esenciales para:

- Manipulación segura de JSON no estándar
- Lectura confiable de archivos del sistema Linux
- Procesamiento de números provenientes de texto
- Robustez total para los módulos que dependen de datos externos

Es un componente fundamental para la estabilidad del sistema de monitorización, garantizando que los datos siempre sean válidos, limpios y utilizables.


