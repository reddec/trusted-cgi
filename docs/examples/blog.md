---
layout: default
title: Blog
parent: Examples
nav_order: 1
---
# Blog example

This is basic public comment board example: 

* anyone can post comment with any name
* only last N post will be displayed
* comments stored in SQLite

**Do not use the example in production!** Comments board without authorization it's a bad, very bad idea.
However, the example should be safe for XSS attacks due to user data escaping.

* Create new project based on python template (UI-> Dashboard -> Python)
* Click files, then click on app.py file. Copy following content:

```python
import os
import sqlite3
import html

# how many last comments to show
last_num = 20
# db location
db_name = 'comments.db'
# define out database schema
schema = '''
CREATE TABLE IF NOT EXISTS comment(
    id      INT PRIMARY KEY,
    name    VARCHAR(100) NOT NULL,
    comment VARCHAR(255) NOT NULL
)
'''
# html template
html_template = '''<html><body>
<div>
<form method="post" enctype="application/x-www-form-urlencoded">
<input type="text" placeholder="your name" name="name" maxlength="100"/><br/>
<textarea placeholder="comment" maxlength="255" name="comment"></textarea><br/>
<button type="submit">send</button>
</form>
</div>
{comments}
</body></html>'''

# form values
name = os.getenv('NAME')  
comment = os.getenv('COMMENT') 

# connect to local database
with sqlite3.connect(db_name) as conn:
    cur = conn.cursor()
    cur.execute(schema)
    cur.execute('INSERT INTO comment (name, comment) VALUES (?, ?)', (name, comment))
    conn.commit()
    latest = cur.execute(f'SELECT id, name, comment FROM comment ORDER BY id DESC LIMIT {last_num}')
    # re-render main page comment block
    blocks = []
    for (id, name, comment) in latest:
        blocks.append(f'<div id="{id}"><p>From: <b>{html.escape(name)}</b></p><p>{html.escape(comment)}</p></div><hr/>')
    page = html_template.replace('{comments}', "\n".join(blocks))
    with open('static/index.html', 'wt') as f:
        f.write(page)
```

* In UI create directory `static`, click on it, create file `index.html` and put following content:

```html
<html><body>
<div>
<form method="post" enctype="application/x-www-form-urlencoded">
<input type="text" placeholder="your name" name="name" maxlength="100"/><br/>
<textarea placeholder="comment" maxlength="255" name="comment"></textarea><br/>
<button type="submit">send</button>
</form>
</div>
</body></html>
```

* Setup mapping for form values: click on Mapping tab and add:
  * output headers: `Content-Type: text/html` 
  * query params should be mapped as: comment to `COMMENT` and name to `NAME`
  * set static dir to `static`
  * click 'save'

![image](https://user-images.githubusercontent.com/6597086/83414645-6e1db100-a450-11ea-9ded-2e5d93ac00f7.png)


Done!

Now open URL from overview section by browser and try to post something.

