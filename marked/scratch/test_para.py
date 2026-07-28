import marked
lexer = marked.Lexer()
print(repr(lexer.tokenizer.rules.block.paragraph.match("fence\n```").group(0)))
