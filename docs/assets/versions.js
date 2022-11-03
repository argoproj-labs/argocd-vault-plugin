setTimeout(function() {
  const callbackName = 'callback_' + new Date().getTime();
  window[callbackName] = function (response) {
    const div = document.createElement('div');
    div.classList.add('md-version')
    div.innerHTML = response.html;
    document.querySelector(".md-header__topic").appendChild(div);
    const container = div.querySelector('.rst-versions');
    container.classList.remove("rst-badge");
    var caret = document.createElement('div');
    caret.innerHTML = "<i class='fa fa-caret-down dropdown-caret'></i>"
    caret.classList.add('dropdown-caret')
    div.querySelector('.rst-current-version').appendChild(caret);

    var items = document.querySelectorAll('.rst-versions .rst-other-versions dd a');
    for (var i = 0, item; item = items[i]; i++) {
      item.setAttribute('id','rtd-version-item');
    }

    var currentVersion = document.getElementsByClassName("rst-current-version");
    currentVersion[0].innerText = currentVersion[0].innerText.split(':')[1].trim()
  }

  var script = document.createElement('script');
  script.src = 'https://argocd-vault-plugin.readthedocs.io/_/api/v2/footer_html/?'+
      'callback=' + callbackName + '&project=argocd-vault-plugin&page=&theme=mkdocs&format=jsonp&docroot=docs&source_suffix=.md&version=' + (window['READTHEDOCS_DATA'] || { version: 'latest' }).version;
  document.getElementsByTagName('head')[0].appendChild(script);
}, 0);

// VERSION WARNINGS
window.addEventListener("DOMContentLoaded", function() {
  var rtdData = window['READTHEDOCS_DATA'] || { version: 'latest' };
  var margin = 30;
  var headerHeight = document.getElementsByClassName("md-header")[0].offsetHeight;
  if (rtdData.version === "latest") {
    document.querySelector("div[data-md-component=announce]").innerHTML = "<div id='announce-msg'>You are viewing the docs for an unreleased version of argocd-vault-plugin, <a href='https://argocd-vault-plugin.readthedocs.io/en/stable/'>click here to go to the latest stable version.</a></div>"
    var bannerHeight = document.getElementById('announce-msg').offsetHeight + margin
    document.querySelector("header.md-header").style.top = bannerHeight +"px";
    document.querySelector('style').textContent +=
    "@media screen and (min-width: 76.25em){ .md-sidebar { height: 0;  top:"+ (bannerHeight+headerHeight)+"px !important; }}"
    document.querySelector('style').textContent +=
    "@media screen and (min-width: 60em){ .md-sidebar--secondary { height: 0;  top:"+ (bannerHeight+headerHeight)+"px !important; }}"
  }
  else if (rtdData.version !== "stable") {
    document.querySelector("div[data-md-component=announce]").innerHTML = "<div id='announce-msg'>You are viewing the docs for a previous version of argocd-vault-plugin, <a href='https://argocd-vault-plugin.readthedocs.io/en/stable/'>click here to go to the latest stable version.</a></div>"
    var bannerHeight = document.getElementById('announce-msg').offsetHeight + margin
    document.querySelector("header.md-header").style.top = bannerHeight +"px";
    document.querySelector('style').textContent +=
    "@media screen and (min-width: 76.25em){ .md-sidebar { height: 0;  top:"+ (bannerHeight+headerHeight)+"px !important; }}"
    document.querySelector('style').textContent +=
    "@media screen and (min-width: 60em){ .md-sidebar--secondary { height: 0;  top:"+ (bannerHeight+headerHeight)+"px !important; }}"
  }
});
