(function () {

  $(function () {
    const $editor = $('#metrics-viewer');

    const $timeSelect = $('<select class="window-select"></select>');
    ['hour', 'day', 'week'].forEach((range) => {
      const $option = $('<option></option>', { value: range });
      $option.text(range[0].toUpperCase() + range.substring(1));
      $timeSelect.append($option);
    });
    $editor.append($timeSelect);

    const $details = $('<div class="details"></div>');
    $editor.append($details);

    function showData() {
      const windowUsage = window.metricsData[$timeSelect.val()];
      const totals = { egress: 0, ingress: 0, requests: 0 };
      const hostCounts = [];
      for (const host of Object.keys(windowUsage)) {
        const hostMetrics = windowUsage[host];
        hostCounts.push([host, hostMetrics]);
        for (const metric of Object.keys(hostMetrics)) {
          totals[metric] += hostMetrics[metric] || 0;
        }
      }
      hostCounts.sort((a, b) => (a[1]['requests'] || 0) - (b[1]['requests'] || 0));
      const $totalsTable = $(`
        <table class="totals">
          <tr><td class="field">Egress</td><td>${totals['egress']}</td></tr>
          <tr><td class="field">Ingress</td><td>${totals['ingress']}</td></tr>
          <tr><td class="field">Requests</td><td>${totals['requests']}</td></tr>
        </table>
      `);
      $details.empty();
      $details.append($totalsTable);

      $hostTable = $('<table class="by-host"><tr><th>Host</th><th>Requests</th><th>Ingress</th><th>Egress</th></tr></table>');
      hostCounts.forEach((hostAndMetrics) => {
        const [host, metrics] = hostAndMetrics;
        const $row = $(`
          <tr>
            <td class="host"></td>
            <td>${metrics['requests'] || 0}</td>
            <td>${metrics['ingress'] || 0}</td>
            <td>${metrics['egress'] || 0}</td>
          </tr>`
        );
        $row.find('.host').text(host);
        $hostTable.append($row);
      });
      $details.append($hostTable);
    }

    $timeSelect.change(showData);
    showData();
  });

})();
