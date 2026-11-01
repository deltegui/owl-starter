# owl-starter

Plantilla de aplicación con [OWL](https://github.com/deltegui/owl) y [Hix](https://github.com/deltegui/hx):
Postgres, sesiones en base de datos, login con CSRF, es/en y Bootstrap (`hxbs`).

```sh
gonew github.com/deltegui/owl-starter example.com/miapp
cd miapp
make init    # crea conf.json (no se sube a git) con un secreto aleatorio
make db      # Postgres en docker
make run     # http://localhost:8080  (admin@example.com / change-me-please)
```

## Estructura

| Carpeta | Qué va |
|---|---|
| `cmd/app` | El cableado: configuración, base de datos, OWL y la lista de features. |
| `feature/<slice>` | Una pantalla o mantenimiento: `_bootstrap`, `_handler`, `_service`, `_locale`. **Un slice no importa a otro.** |
| `domain` | Lo que comparten varias features (tipos y contratos de negocio). Sin SQL. |
| `sys` | La "fontanería" del sistema: `config`, `bootstrap` (lo que recibe cada feature) y `persistence` (Postgres, migraciones, repositorios y stores). |
| `web` | Lo común de la interfaz: `layout`, `locale` (textos compartidos) y `static`. |

Los textos de una pantalla son un `localizer.Dict` junto a su feature.

## Configuración

Con [configloader](https://github.com/deltegui/configloader) v3, en este orden (gana la última):
valores por defecto, `conf.json` (opcional, plantilla en `conf.example.json`), variables `APP_*` y flags (`-h` los lista).
Si falta algo obligatorio (`APP_SECRET`, `APP_DATABASE_URL`) o es inválido, la app **no arranca** y
dice todos los problemas a la vez. Los campos están en `sys/config/config.go`.

## Base de datos

Las migraciones son ficheros de `sys/persistence/migrations/` que se aplican al arrancar, en orden y
una vez. Para cambiar el esquema añade `002_algo.sql`; no edites uno ya aplicado.
