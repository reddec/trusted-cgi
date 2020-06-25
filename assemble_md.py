import os
import sys
import argparse


parser = argparse.ArgumentParser(description='Assembe markdown in one doc')
parser.add_argument('root_dir', type=str, help='dir to scan')
parser.add_argument('--exclude', type=str, nargs='*',  help='paths to exclude')
args = parser.parse_args()


def add_level(line, depth):
    if not line.startswith('#'):
        return line
    num = 0
    title = ''
    for i, x in enumerate(line):
        if x == '#':
            num += 1
        else:
            title = line[i:]
            break
    return '#' * (depth + num) + title

root_dir = args.root_dir
exclude = set(args.exclude or [])
for root, dirs, files in os.walk(root_dir):
    ignore = False
    for dir in exclude:
        if root.startswith(dir):
            ignore = True
            break
    if ignore:
        continue

    title = ''
    root_depth = root[len(root_dir):].count(os.sep)

    if root_depth == 0:
        title = 'Manual'
    else:
        title = os.path.basename(root)

    print()
    print("#" * root_depth, title.title().replace('_',' '))
    print()

    root_depth += 1

    if 'index.md' in files:
        idx = files.index('index.md')
        files[0], files[idx] = files[idx], files[0]

    for file in files:
        if not file.lower().endswith('.md'):
            continue
        depth = root_depth
        section = file[:-3].strip()
        if section == 'index':
            depth += 1
        else:
            print()
            print("#" * depth, section.title().replace('_',' '))
            print()
        with open(os.path.join(root, file), 'rt') as f:
            line = next(f)
            if line.startswith('-'):
                for line in f:
                    if line.startswith('-'):
                        break
            else:
                print(add_level(line,depth))
            for line in f:
                print(add_level(line,depth))
            


    

        