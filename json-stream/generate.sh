mkdir -p generated
mkdir -p generated/lexers
mkdir -p generated/parsers

go run github.com/johnkerl/pgpg/generators/go/cmd/lexgen-tables \
  -o generated/lexers/mylexer.json \
  json.bnf

go run github.com/johnkerl/pgpg/generators/go/cmd/lexgen-code \
  -o generated/lexers/mylexer.go \
  -package lexers \
  -type MyLexer \
  generated/lexers/mylexer.json

go run github.com/johnkerl/pgpg/generators/go/cmd/parsegen-tables \
  -o generated/parsers/myparser.json \
  json.bnf

go run github.com/johnkerl/pgpg/generators/go/cmd/parsegen-code \
  -o generated/parsers/myparser.go \
  -package parsers \
  -type MyParser \
  generated/parsers/myparser.json
