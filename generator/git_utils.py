import subprocess
import os

GIT_ENV = os.environ.copy()
GIT_ENV['GIT_CONFIG_GLOBAL'] = '/dev/null'
GIT_ENV['GIT_AUTHOR_NAME'] = 'Alex Vance'
GIT_ENV['GIT_AUTHOR_EMAIL'] = 'alex.vance@kestrelflow.io'
GIT_ENV['GIT_COMMITTER_NAME'] = 'Alex Vance'
GIT_ENV['GIT_COMMITTER_EMAIL'] = 'alex.vance@kestrelflow.io'

def commit(message, date_str, files=None):
    env = GIT_ENV.copy()
    env['GIT_AUTHOR_DATE'] = date_str
    env['GIT_COMMITTER_DATE'] = date_str
    
    if files:
        for f in files:
            subprocess.run(['git', 'add', f], env=env, check=True)
    else:
        subprocess.run(['git', 'add', '-A'], env=env, check=True)
        
    res = subprocess.run(['git', 'status', '--porcelain'], env=env, capture_output=True, text=True)
    if not res.stdout.strip():
        # Nothing to commit
        return False
        
    subprocess.run(['git', 'commit', '-m', message], env=env, check=True)
    return True

def get_commit_count():
    res = subprocess.run(['git', 'rev-list', '--count', 'HEAD'], env=GIT_ENV, capture_output=True, text=True)
    return int(res.stdout.strip()) if res.returncode == 0 else 0

def tag(tag_name, message=None):
    cmd = ['git', 'tag', tag_name]
    if message:
        cmd.extend(['-m', message])
    subprocess.run(cmd, env=GIT_ENV, check=True)
