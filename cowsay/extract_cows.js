const fs = require('fs');
const spec = fs.readFileSync('spec/SPEC.md', 'utf8');
const cowsDir = 'target/cows/';
fs.mkdirSync(cowsDir, { recursive: true });

const regex = /### ([^\n]+)\n```perl\n([\s\S]*?)\n```/g;
let match;
while ((match = regex.exec(spec)) !== null) {
  const name = match[1].trim();
  const content = match[2];
  fs.writeFileSync(cowsDir + name + '.cow', content);
}
console.log('Done extracting cows.');
