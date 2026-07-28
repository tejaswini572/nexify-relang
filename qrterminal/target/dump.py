data = b'https://example.com'
# byte mode = 0100
# len = 19 (00010011)
bits = '0100' + format(19, '08b')
for b in data:
    bits += format(b, '08b')
# terminator
cap = 34 * 8
bits += '0' * min(4, cap - len(bits))
# pad to byte
while len(bits) % 8 != 0:
    bits += '0'
# pad with ec 11
pad = ['11101100', '00010001']
i = 0
while len(bits) < cap:
    bits += pad[i%2]
    i += 1
bytes_list = [int(bits[j:j+8], 2) for j in range(0, len(bits), 8)]
print('Data:', ' '.join(f'{b:02x}' for b in bytes_list))
# EC calculation
def mul(a, b):
    if a==0 or b==0: return 0
    r = 0
    for i in range(8):
        if (b & 1) != 0: r ^= a
        hi = a & 0x80
        a = (a << 1) & 0xFF
        if hi != 0: a ^= 0x1d
        b >>= 1
    return r
g = [1]
root = 1
for _ in range(10):
    new_g = [0]*(len(g)+1)
    for i in range(len(g)): new_g[i] = g[i]
    for i in range(len(g)):
        if g[i] != 0: new_g[i+1] ^= mul(g[i], root)
    g = new_g
    root = mul(root, 2)
msg = list(bytes_list) + [0]*10
for i in range(len(bytes_list)):
    coef = msg[i]
    if coef != 0:
        for j in range(len(g)):
            msg[i+j] ^= mul(g[j], coef)
print('EC:', ' '.join(f'{msg[i]:02x}' for i in range(len(bytes_list), len(msg))))
