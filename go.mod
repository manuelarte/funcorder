module github.com/manuelarte/funcorder

go 1.24.0

require (
	github.com/dave/dst v0.27.3
	github.com/pmezard/go-difflib v1.0.0
	golang.org/x/tools v0.38.0
)

require (
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)

retract v0.4.0 // Major bug found when introducing suggested fixes, issue #32
