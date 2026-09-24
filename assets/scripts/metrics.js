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
      const byDay = window.metricsData[$timeSelect.val()];
      console.log('by day', byDay, $timeSelect.val());
      const totals = { egress: 0, ingress: 0, requests: 0 };
      for (const host of Object.keys(byDay)) {
        for (const metric of Object.keys(byDay[host])) {
          totals[metric] += byDay[host][metric] || 0;
        }
      }
      const $totalsTable = $(`
        <table>
          <tr><td class="field">Egress</td><td>${totals['egress']}</td></tr>
          <tr><td class="field">Ingress</td><td>${totals['ingress']}</td></tr>
          <tr><td class="field">Requests</td><td>${totals['requests']}</td></tr>
        </table>
      `);
      $details.empty();
      $details.append($totalsTable);
    }

    $timeSelect.change(showData);
    showData();
  });

})();
