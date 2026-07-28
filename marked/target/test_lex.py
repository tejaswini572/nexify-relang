import marked
import json
lexer = marked.Lexer()
tokens = lexer.lex('fence\n```')
print(json.dumps(tokens, indent=2))
