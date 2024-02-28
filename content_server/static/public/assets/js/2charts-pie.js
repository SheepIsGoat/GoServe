function configPieChart(pieConfig) {
    const pieCtx = document.getElementById('pie');
    window.myPie = new Chart(pieCtx, pieConfig);
}