module tryparse

go 1.25

require github.com/johnkerl/pgpg/lib v0.0.0

replace github.com/johnkerl/pgpg/generated => ../../pgpg/apps/go/generated

replace github.com/johnkerl/pgpg/generators/go => ../../pgpg/generators/go

replace github.com/johnkerl/pgpg/lib => ../../pgpg/lib
