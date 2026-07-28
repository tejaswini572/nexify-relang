import sys
from rain import Matrix
matrix = Matrix(screen_width=10, line_count=2, line_speed=0)

matrix._setScreenLineArray()
matrix.line_array[0] = 0 # empty
matrix.line_array[1] = 1 # char
matrix.line_array[2] = 0 # empty
matrix.line_array[3] = 1 # char

line = ""
for m in range(4):
    n = matrix.line_array[m]
    if n == 1:
        line = line + matrix._getTextColourRandomChar() + matrix._getCharacter()
    else:
        line = line + matrix._getTextColourRandomChar() + " "
print(repr(line))
