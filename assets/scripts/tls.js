(function() {

  function TlsEditor(config) {
    var tlsConfig = config.tlsConfig;
    var redirects = (config.redirects || []);
    this._$mainContent = $('#tls-settings');
    this._$default = generateKeyCert(tlsConfig.default.key, tlsConfig.default.certificate);
    this._$mainContent.append('<h1 class="field-set-heading">Default Key/Cert Pair</h1>',
      this._$default);
    this._initializeRootCAs(tlsConfig);
    this._initializeNamedCertificates(tlsConfig);
    this._initializeRedirects(redirects);
    this._initializeACME(tlsConfig);
    this._fixFloatingTextareaBug();
  }

  TlsEditor.prototype.getConfig = function() {
    var rootCAs = [];
    $('.root-ca textarea').each(function(index, element) {
      rootCAs.push($(element).val());
    });
    var named = {};
    $('.named-key-cert-pair').each(function(index, element) {
      var $element = $(element);
      var name = $element.find('.key-cert-name').val();
      var key = $element.find('.key-value').val();
      var cert = $element.find('.cert-value').val();
      named[name] = {key: key, certificate: cert};
    });
    var redirects = [];
    $('.https-redirect-host input').each(function(index, element) {
      redirects.push($(element).val());
    });
    var acmeHosts = [];
    $('.acme-host input').each(function(index, element) {
      acmeHosts.push($(element).val());
    });
    return {
      tlsConfig: {
        default: {
          key: this._$default.find('.key-value').val(),
          certificate: this._$default.find('.cert-value').val()
        },
        root_ca: rootCAs,
        named: named,
        acme_dir_url: this._$mainContent.find('.acme-directory-url').val(),
        acme_hosts: acmeHosts,
      },
      redirects: redirects
    };
  };

  TlsEditor.prototype._fixFloatingTextareaBug = function() {
    // In Safari and Chrome as of July 15, 2015, floating <textarea>'s disappear when the text
    // scrolls and then the user deletes the text.
    var $textareas = $('.textarea-field-textarea');
    $textareas.each(function(index, element) {
      var $element = $(element);
      var elementStyle = $element.css(['float', 'width', 'height', 'box-sizing', 'display']);
      var $parent = $('<div></div>').css(elementStyle);
      $element.css({
        width: '100%',
        height: '100%',
        boxSizing: 'border-box',
        float: 'none'
      });
      $element.after($parent);
      $element.detach();
      $parent.append($element);
    });
  };

  TlsEditor.prototype._initializeNamedCertificates = function(tlsConfig) {
    var $heading = $('<div class="field-set-action-heading"><h1>Certificates</h1>' +
      '<button class="field-set-add-button">Add</button></div>');
    var $certs = $('<div class="named-certificates"></div>');
    var keys = Object.keys(tlsConfig.named).sort();
    for (var i = 0, len = keys.length; i < len; ++i) {
      var name = keys[i];
      var kc = tlsConfig.named[name];
      $certs.append(generateNamedKeyCert(name, kc.key, kc.certificate));
    }
    $heading.find('button').click(function() {
      $certs.prepend(generateNamedKeyCert('', '', ''));
    }.bind(this));
    this._$certs = $certs;
    this._$mainContent.append($heading, $certs);
  };

  TlsEditor.prototype._initializeRootCAs = function(tlsConfig) {
    var $heading = $('<div class="field-set-action-heading"><h1>Root CAs</h1>' +
      '<button class="field-set-add-button">Add</button></div>');
    var $cas = $('<div class="root-cas"></div>');
    for (var i = 0, len = tlsConfig.root_ca.length; i < len; ++i) {
      var ca = tlsConfig.root_ca[i];
      $cas.append(generateRootCA(ca));
    }
    $heading.find('button').click(function() {
      $cas.prepend(generateRootCA(''));
    }.bind(this));
    this._$cas = $cas;
    this._$mainContent.append($heading, $cas);
  };

  TlsEditor.prototype._initializeRedirects = function(redirects) {
    var $listTitle = $('<div class="field-set-action-heading">' +
      '<h1>Redirects</h1><button class="field-set-add-button">Add</button></div>');

    redirects.sort();
    var $redirects = $('<div></div>');
    for (var i = 0; i < redirects.length; ++i) {
      $redirects.append(generateRedirect(redirects[i]));
    }

    $listTitle.find('.field-set-add-button').click(function() {
      $redirects.prepend(generateRedirect(''));
    });

    this._$mainContent.append($listTitle, $redirects);
  };

  TlsEditor.prototype._initializeACME = function(tlsConfig) {
    var $dirURL = $('<div class="field">' +
      '<label class="input-field-label">ACME Directory</label>' +
      '<input class="input-field-input acme-directory-url"></div>');
    var $urlInput = $dirURL.find('input');
    $urlInput.val(tlsConfig.acme_dir_url);
    $urlInput.attr('placeholder', 'Defaults to LetsEncrypt');

    var $listTitle = $('<div class="field-set-action-heading">' +
      '<h1>ACME Hosts</h1><button class="field-set-add-button">Add</button></div>');

    var hosts = (tlsConfig.acme_hosts || []).slice();
    hosts.sort();
    var $hosts = $('<div></div>');
    for (var i = 0; i < hosts.length; ++i) {
      $hosts.append(generateACMEHost(hosts[i]));
    }

    $listTitle.find('.field-set-add-button').click(function() {
      $hosts.prepend(generateACMEHost(''));
    });

    $acmeSettings = $('<div></div>').append($dirURL, $listTitle, $hosts);
    $acmeSettings.addClass('acme-settings');

    this._$mainContent.append($acmeSettings);
  };

  function generateKeyCert(key, cert) {
    var $res = $('<div class="key-cert-pair"><div class="field">' +
      '<label class="textarea-field-label">Key</label>' +
      '<textarea class="textarea-field-textarea key-value"></textarea>' +
      '</div><div class="field">' +
      '<label class="textarea-field-label">Certificate</label>' +
      '<textarea class="textarea-field-textarea cert-value"></textarea></div>');
    $res.find('.key-value').val(key);
    $res.find('.cert-value').val(cert);
    return $res;
  }

  function generateNamedKeyCert(name, key, cert) {
    var $res = generateKeyCert(key, cert);
    $res.addClass('named-key-cert-pair');
    $res.prepend('<div class="field"><label class="input-field-label">Name</label>' +
      '<input class="input-field-input key-cert-name"></div>');
    $res.append('<button class="unlabeled-field">Delete</button>');
    $res.find('.key-cert-name').val(name);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function generateRootCA(ca) {
    var $res = $('<div class="root-ca"><textarea></textarea>' +
      '<button>Delete</button></div>');
    $res.find('textarea').val(ca);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function generateRedirect(hostname) {
    var $res = $('<div class="https-redirect-host"><input placeholder="Host">' +
      '<button>Remove</button></div>');
    $res.find('input').val(hostname);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function generateACMEHost(hostname) {
    var $res = $('<div class="acme-host"><input placeholder="Host">' +
      '<button>Remove</button></div>');
    $res.find('input').val(hostname);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  $(function() {
    var editor = new TlsEditor(window.tlsConfiguration);
    $('#submit').click(function() {
      var rulesJSON = JSON.stringify(editor.getConfig());
      postData('rules', rulesJSON, '/set_tls');
    });
  });

})();
