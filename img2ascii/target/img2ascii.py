#!/usr/bin/env python3
import sys
import getopt
import math
from PIL import Image

GRAYSCALE_FLAG = 1 << 0
REVERSE_FLAG   = 1 << 1
PRINT_FLAG     = 1 << 2
DEBUG_FLAG     = 1 << 3

def show_usage():
    sys.stdout.write(
        "\nUsage: \x1b[1mimg2ascii [options] -i <FILE> [-o <FILE>]\x1b[0m \n\n"
        "A command-line tool for converting images to ASCII art \n\n"
        "Options: \n"
        "   -i, --input  <FILE>     Path of the input image file (required) \n"
        "   -o, --output <FILE>     Path of the output file \n"
        "   -w, --width  <NUMBER>   Width of the output \n"
        "   -c, --chars  <STRING>   Characters to be used for the ASCII image \n"
        "   -p, --print             Print the output to the console \n"
        "   -r, --reverse           Reverse the string of characters \n"
        "   -d, --debug             Print some useful information \n\n"
    )

def c_round(val):
    return int(math.floor(val + 0.5))

def reverse_string(s):
    return s[::-1]

def get_intensity(r, g, b):
    return min(255, max(0, c_round(0.299 * r + 0.587 * g + 0.114 * b)))

def get_output_grayscale(pixels, width, height, characters, flags):
    if width <= 0 or height <= 0 or pixels is None:
        return ""
    if flags & REVERSE_FLAG:
        characters = reverse_string(characters)
    
    char_len = len(characters)
    output = []
    
    for y in range(height):
        for x in range(width):
            r, g, b = pixels[x, y]
            intensity = get_intensity(r, g, b)
            char_index = int(intensity / (255.0 / (char_len - 1)))
            output.append(characters[char_index])
        output.append('\n')
        
    return "".join(output)

def get_output_rgb(pixels, width, height, characters, flags):
    if width <= 0 or height <= 0 or pixels is None:
        return "\x1b[0m"
    if flags & REVERSE_FLAG:
        characters = reverse_string(characters)
        
    char_len = len(characters)
    output = []
    
    r_prev, g_prev, b_prev = -1, -1, -1
    
    for y in range(height):
        for x in range(width):
            r, g, b = pixels[x, y]
            intensity = get_intensity(r, g, b)
            char_index = int(intensity / (255.0 / (char_len - 1)))
            
            if not (r == r_prev and g == g_prev and b == b_prev):
                output.append(f"\x1b[38;2;{r};{g};{b}m")
                r_prev, g_prev, b_prev = r, g, b
                
            output.append(characters[char_index])
        output.append('\n')
        
    output.append("\x1b[0m")
    return "".join(output)

def main():
    if len(sys.argv) == 1:
        sys.stdout.write("No input file\n")
        show_usage()
        sys.exit(1)
        
    try:
        opts, args = getopt.getopt(
            sys.argv[1:],
            "hi:o:w:c:gprd",
            ["help", "input=", "output=", "width=", "chars=", "grayscale", "print", "reverse", "debug"]
        )
    except getopt.GetoptError as err:
        sys.stdout.write(str(err) + "\n")
        sys.stdout.write("\nHint: Use the \x1b[1m--help\x1b[0m option to get help about the usage \n\n")
        sys.exit(1)
        
    input_filepath = None
    output_filepath = None
    characters = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'. "
    desired_width = 0
    flags = 0
    resize_image = False
    
    for o, a in opts:
        if o in ("-h", "--help"):
            show_usage()
            sys.exit(1)
        elif o in ("-i", "--input"):
            input_filepath = a
        elif o in ("-o", "--output"):
            output_filepath = a
        elif o in ("-w", "--width"):
            desired_width = int(a)
            resize_image = True
        elif o in ("-c", "--chars"):
            if len(a) > 0:
                characters = a
        elif o in ("-g", "--grayscale"):
            flags |= GRAYSCALE_FLAG
        elif o in ("-p", "--print"):
            flags |= PRINT_FLAG
        elif o in ("-r", "--reverse"):
            flags |= REVERSE_FLAG
        elif o in ("-d", "--debug"):
            flags |= DEBUG_FLAG
            
    if input_filepath is None:
        sys.stdout.write("No input file\n")
        show_usage()
        sys.exit(1)
        
    if output_filepath is None:
        flags |= PRINT_FLAG
        
    # Load and resize the image
    try:
        image = Image.open(input_filepath).convert("RGB")
    except Exception as e:
        sys.stderr.write(f"Could not load image: {e}\n")
        sys.exit(1)
        
    width, height = image.size
    
    if resize_image:
        if desired_width <= 0:
            sys.stderr.write("Argument 'width' must be greater than 0 \n")
            sys.exit(1)
        elif desired_width > width:
            sys.stderr.write(f"Argument 'width' can not be greater than the original image width ({width}px) \n")
            sys.exit(1)
            
        desired_height = int(height / (width / float(desired_width)) / 2)
    else:
        desired_width = width
        desired_height = height // 2
        
    if desired_width > 0 and desired_height > 0:
        image = image.resize((desired_width, desired_height), Image.Resampling.BILINEAR)
        pixels = image.load()
    else:
        pixels = None
        
    # Generate Output
    if flags & GRAYSCALE_FLAG:
        output = get_output_grayscale(pixels, desired_width, desired_height, characters, flags)
    else:
        output = get_output_rgb(pixels, desired_width, desired_height, characters, flags)
        
    if flags & DEBUG_FLAG:
        sys.stdout.write(
            f"Input: {input_filepath} \n"
            f"Output: {output_filepath if output_filepath is not None else 'stdout'} \n"
            f"Resolution: {desired_width}x{desired_height} \n"
            f"Characters ({len(characters)}): \"{characters}\" \n"
        )
        
    if flags & PRINT_FLAG:
        sys.stdout.write(output)
        
    if output_filepath is not None:
        try:
            with open(output_filepath, "w", encoding="utf-8") as f:
                f.write(output)
        except Exception as e:
            sys.stderr.write(f"Could not create an output file: {e}\n")
            sys.exit(1)

if __name__ == "__main__":
    main()
