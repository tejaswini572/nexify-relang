import('../source/lib/marked.esm.js').then(m => {
  const rules = m.marked.Lexer.rules;
  const inline = rules.inline;
  console.log('inline keys:', Object.keys(inline));
  for (const mode of ['normal', 'gfm', 'breaks', 'pedantic']) {
    const r = inline[mode];
    if (!r) continue;
    console.log(`\n=== ${mode} ===`);
    for (const [k, v] of Object.entries(r)) {
      if (v && v.source) {
        if (k.includes('em') || k.includes('del') || k.includes('Del')) {
          console.log(`${k}: /${v.source}/${v.flags}`);
        }
      }
    }
  }
});
