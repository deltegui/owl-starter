module github.com/deltegui/owl-starter

go 1.25.0

require (
	github.com/deltegui/configloader/v3 v3.0.0
	github.com/deltegui/hx v0.3.0
	github.com/deltegui/hxbs v0.3.0
	github.com/deltegui/owl/v3 v3.0.0-rc.5
	github.com/jackc/pgx/v5 v5.7.6
)

require (
	github.com/deltegui/valtruc v1.4.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

// While OWL is a release candidate this points at the local copy.
replace github.com/deltegui/owl/v3 => ../owl
