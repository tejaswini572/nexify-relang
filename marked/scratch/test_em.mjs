import('../source/lib/marked.esm.js').then(m => {
  const tests = [
    '**bold ~~strike and *italic*~~ bold**',
  ];
  for (const t of tests) {
    console.log('INPUT:', JSON.stringify(t));
    console.log('OUTPUT:', JSON.stringify(m.marked.parse(t)));
    console.log('TOKENS:', JSON.stringify(m.marked.lexer(t)[0].tokens, null, 2));
    console.log('---');
  }
});
