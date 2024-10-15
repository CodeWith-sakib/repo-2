import os

def count_go_file(filepath):
    in_block = False
    blanks = 0
    comments = 0
    code = 0
    with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
        for line in f:
            stripped = line.strip()
            if not stripped:
                blanks += 1
                continue
            if in_block:
                comments += 1
                if '*/' in stripped:
                    in_block = False
                continue
            if stripped.startswith('/*'):
                comments += 1
                if '*/' not in stripped:
                    in_block = True
                continue
            if stripped.startswith('//'):
                comments += 1
                continue
            code += 1
    return blanks, comments, code

def get_production_loc(repo_root='.'):
    total_code = 0
    total_comments = 0
    total_blanks = 0
    file_counts = {}
    
    for root, dirs, files in os.walk(repo_root):
        # Exclude git, vendor, internal-bench, scripts, generator
        dirs[:] = [d for d in dirs if d not in ('.git', 'vendor', 'internal-bench', 'generator', 'scripts', '.github')]
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go'):
                path = os.path.join(root, f)
                rel_path = os.path.relpath(path, repo_root)
                blanks, comments, code = count_go_file(path)
                total_blanks += blanks
                total_comments += comments
                total_code += code
                file_counts[rel_path] = code
                
    return {
        'code': total_code,
        'comments': total_comments,
        'blanks': total_blanks,
        'total_lines': total_code + total_comments + total_blanks,
        'files': file_counts
    }

if __name__ == '__main__':
    res = get_production_loc('.')
    print(f"Production Go LOC: {res['code']} across {len(res['files'])} files")
