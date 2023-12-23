---
title: 'Hugo Whisper'
date: 2019-02-11T19:27:37+10:00
weight: 6
---

Whisper is a minimal documentation theme built for Hugo. The design &amp; functionality is intentionally minimal.

<!--more-->

## Quickstart

Copy or git clone this theme into the sites themes folder `mynewsite/themes`

```
hugo new site whisper
git clone https://github.com/jugglerx/hugo-whisper-theme.git
```

### Code Highlighting

Whisper uses Hugo's in-built code highlighting with a github style code highlighting theme. https://gohugo.io/content-management/syntax-highlighting/

You can insert code snippets in any markdown file by using standard code fences syntax ie:

```js
function myFunction() {
  var x = document.getElementById('myDIV');
  if (x.style.display === 'none') {
    x.style.display = 'block';
  } else {
    x.style.display = 'none';
  }
}
```
