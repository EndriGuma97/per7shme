# ePer7shme Jobs

Faqe moderne rekrutimi e ndërtuar me **Go + HTMX + SQLite** — pa framework të rëndë, pa Node, pa CGO.

## Nisja

```bash
go run .
```

Faqja hapet në <http://localhost:8080> (porta ndryshohet me variablin `PORT`).

Në nisjen e parë krijohet automatikisht databaza `eper7shme.db` me:

- një administrator — **admin / eper7shme2026** (ndryshohet me variablat `ADMIN_USER` / `ADMIN_PASSWORD` përpara nisjes së parë)
- 3 postime shembull në blog

## Paneli i administrimit

Hyni te <http://localhost:8080/admin> për të krijuar, ndryshuar dhe fshirë postimet e blogut.
Përmbajtja shkruhet në **Markdown** (tituj `##`, **bold**, lista…).

## Struktura

| Skedari | Roli |
|---|---|
| `main.go` | Rrugët HTTP dhe nisja e serverit |
| `store.go` | SQLite: migrimet, seed, postimet, sesionet |
| `handlers.go` | Faqet publike (kreu, blogu, artikujt) |
| `admin.go` | Hyrja + CRUD i postimeve |
| `templates/` | Shabllonet HTML (publike + admin) |
| `static/css`, `static/js` | Stili dhe animacionet |

## Teknologjitë

- **Go** (stdlib `net/http`, pa framework)
- **SQLite** përmes `modernc.org/sqlite` (Go i pastër, pa CGO)
- **HTMX** — kërkim live në blog, "më shumë artikuj", fshirje pa rifreskim
- **bcrypt** për fjalëkalimet, sesione me cookie HttpOnly
- **goldmark** për Markdown
