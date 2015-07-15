(function() {

  function TlsEditor(tlsConfig) {
    this._$mainContent = $('#tls-editor');
    this._$default = generateKeyCert(tlsConfig.default.key, tlsConfig.default.certificate);
    this._$mainContent.append('<h1 class="field-set-heading">Default Key/Cert Pair</h1>',
      this._$default);
    this._initializeRootCAs(tlsConfig);
    this._initializeNamedCertificates(tlsConfig);
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
    return {
      default: {
        key: this._$default.find('.key-value').val(),
        certificate: this._$default.find('.cert-value').val()
      },
      root_ca: rootCAs,
      named: named
    };
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

  $(function() {
    var editor = new TlsEditor(window.tlsConfiguration);
    $('#submit').click(function() {
      var rulesJSON = JSON.stringify(editor.getConfig());
      var $form = $('<form method="POST" action="/set_tls"><input name="rules" ' +
        'type="hidden"></form>');
      $form.find('input').val(rulesJSON);
      $form.submit();
    });
  });

})();
