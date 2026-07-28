import('../source/lib/marked.esm.js').then(m => {
  const input = '> - ```\n> > code\n> > ```\n> - end';
  console.log('INPUT:', JSON.stringify(input));
  console.log('OUTPUT:', JSON.stringify(m.marked.parse(input)));
});
