mkdir -p generated
mkdir -p generated/lexers
mkdir -p generated/parsers

go run github.com/johnkerl/pgpg/go/generators/cmd/lexgen-tables \
  -o generated/lexers/mylexer.json \
  dkvp.bnf

go run github.com/johnkerl/pgpg/go/generators/cmd/lexgen-code \
  -o generated/lexers/mylexer.go \
  -package lexers \
  -type MyLexer \
  generated/lexers/mylexer.json

go run github.com/johnkerl/pgpg/go/generators/cmd/parsegen-tables \
  -o generated/parsers/myparser.json \
  dkvp.bnf

go run github.com/johnkerl/pgpg/go/generators/cmd/parsegen-code \
  -o generated/parsers/myparser.go \
  -package parsers \
  -type MyParser \
  generated/parsers/myparser.json
