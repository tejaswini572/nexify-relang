import sys
from rain import Matrix
matrix = Matrix(screen_width=10, line_count=2, line_speed=0)

matrix._setScreenLineArray()
for l in range(matrix._line_count):
    line = ""
    for m, n in matrix.line_array.items():
        if n == 1 or n == 2:
            if n == 2:
                line = line + matrix._getTextColourLightGreenChar() + matrix._getCharacter()
                matrix.line_array[m] = 1
            else:
                line = line + matrix._getTextColourRandomChar() + matrix._getCharacter()
            if 1 == __import__('random').randint(1, 30):
                matrix.line_array[m] = 0
        else:
            line = line + matrix._getTextColourRandomChar() + " "
            if 1 == __import__('random').randint(1, 60):
                matrix.line_array[m] = 2
    print(repr(line))
