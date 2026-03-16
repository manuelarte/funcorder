module github.com/manuelarte/funcorder

go 1.25.0

require golang.org/x/tools v0.43.0

require (
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

retract v0.4.0 // Major bug found when introducing suggested fixes, issue #32
