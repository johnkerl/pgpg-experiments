module tryparse

go 1.25

require github.com/johnkerl/pgpg/lib v0.0.0

require github.com/johnkerl/pgpg/generators/go v0.0.0-20260222224658-91b3b74e1175 // indirect

replace github.com/johnkerl/pgpg/generated => ../../pgpg/apps/go/generated

replace github.com/johnkerl/pgpg/generators/go => ../../pgpg/generators/go

replace github.com/johnkerl/pgpg/lib => ../../pgpg/lib
