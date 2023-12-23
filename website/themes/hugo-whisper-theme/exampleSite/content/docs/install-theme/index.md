---
title: 'Install Theme'
date: 2019-02-11T19:27:37+10:00
draft: false
weight: 3
---

## Create a new Hugo site

```
hugo new site mynewsite
```

This will create a fresh Hugo site in the folder `mynewsite`.

## Install theme

Copy or git clone this theme into the sites themes folder `mynewsite/themes`

#### Install with Git

```
cd mynewsite
cd themes
git clone https://github.com/jugglerx/hugo-whisper-theme.git
```

#### Install from .zip file

You can download the .zip file located here https://github.com/JugglerX/hugo-whisper-theme/archive/master.zip.

Extract the downloaded .zip inside the `themes` folder. Rename the extracted folder from `hugo-whisper-theme-master` -> `hugo-whisper-theme`. You should end up with the following folder structure `mynewsite/themes/hugo-whisper-theme`

## Add example content

The fastest way to get started is to copy the example content and modify the included `config.toml`

### Copy exampleSite contents

Copy the entire contents of the `exampleSite` folder to the root folder of your Hugo site _(the folder with the README.md)_.

### Update config.toml

After you copy the `config.toml` into the root folder of your Hugo site you will need to update the `baseURL`, `themesDir` and `theme` values in the `config.toml`

```
baseURL = "/"
themesDir = "themes"
theme = "hugo-whisper-theme"
```

## Run Hugo

After installing the theme for the first time, generate the Hugo site.

```
hugo
```

For local development run Hugo's built-in local server.

```
hugo server
```

Now enter [`localhost:1313`](http://localhost:1313) in the address bar of your browser.
