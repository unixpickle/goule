(function() {

  function TlsEditor(tlsConfig) {
    this._$mainContent = $('#tls-editor');
    this._$default = generateKeyCert(tlsConfig.default.key, tlsConfig.default.certificate);
    this._$mainContent.append('<h1 class="field-set-heading">Default Key/Cert Pair</h1>',
      this._$default);
    this._initializeRootCAs(tlsConfig);
    this._initializeNamedCertificates(tlsConfig);
  }

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

  function generateKeyCert(key, cert) {
    var $res = $('<div class="key-cert-pair"><div class="field">' +
      '<label class="textarea-field-label">Key</label>' +
      '<textarea class="textarea-field-textarea key-value"></textarea>' +
      '</div><div class="field">' +
      '<label class="textarea-field-label">Certificate</label>' +
      '<textarea class="textarea-field-textarea cert-value"></textarea></div>');
    $res.find('.key-value').text(key);
    $res.find('.cert-value').text(cert);
    return $res;
  }

  function generateNamedKeyCert(name, key, cert) {
    var $res = generateKeyCert(key, cert);
    $res.addClass('named-key-cert-pair');
    $res.prepend('<div class="field"><label class="input-field-label">Name</label>' +
      '<input class="input-field-input key-cert-name"></div>');
    $res.append('<button class="unlabeled-field">Delete</button>');
    $res.find('key-cert-name').text(name);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  function generateRootCA(ca) {
    var $res = $('<div class="root-ca"><textarea></textarea>' +
      '<button>Delete</button></div>');
    $res.find('textarea').text(ca);
    $res.find('button').click(function() {
      $res.remove();
    });
    return $res;
  }

  $(function() {
    new TlsEditor(window.tlsConfiguration);
  });

})();
