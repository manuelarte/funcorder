module github.com/manuelarte/funcorder

go 1.26.0

require golang.org/x/tools v0.50.0

require (
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
)

retract v0.4.0 // Major bug found when introducing suggested fixes, issue #32
