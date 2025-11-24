module github.com/manuelarte/funcorder

go 1.24.0

require golang.org/x/tools v0.39.0

require (
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
)

retract v0.4.0 // Major bug found when introducing suggested fixes, issue #32
