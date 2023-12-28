import './PortfolioMapChart.css'
import { Chart } from 'react-google-charts';
import { toPercent, toUSD } from "./Helpers";

// Define the options for the portfolio map drop-down pickers.
export const PortfolioMapSizeSelectOptions = [
    { value: 'marketValue', label: 'Market Value' },
    { value: 'grossMargin', label: 'Gross Margin'},
];
export const PortfolioMapColorSelectOptions = [
    { value: 'grossMargin', label: 'Gross Margin'},
    { value: 'revenueGrowthPercentageYoy', label: 'Growth Rate TTM' },
    { value: 'revenueGrowthPercentageNextYear', label: 'Growth Rate NTM' },
    { value: 'priceToSalesNtm', label: 'Fwd P/S'},
];

export function PortfolioMapChart({ chartData, sizeBy, colorBy }) {

    function getFilteredChartData() {
        // Filter for only non-zero equities.
        let filtered = chartData.filter(s => (s.marketValue > 0.0 && s.equityType === 'Stock'));
        // Put the filtered values in an array, and add a column header.
        let data = filtered.map(s => [s.ticker, s.equityType,
            sizeBy === 'marketValue' ? s[sizeBy] : s[sizeBy] * 100,
            colorBy === 'priceToSalesNtm' ? s[colorBy] : s[colorBy] * 100])
        // Add the column labels we need for the tree parents.
        data.unshift(['Stock', null, 0, 0])
        data.unshift(['Ticker', 'Type', 'Size Col', 'Color Col'])
        return data
    }

    // Filter the map data.
    var data = getFilteredChartData()
    // Determine coloring scheme based on selection.
    var minColor = 'red'
    var midColor = 'grey'
    var maxColor = 'green'
    var minColorVal = 0
    var maxColorVal = 1
    if (colorBy === 'grossMargin') {
        minColorVal = 0
        maxColorVal = 90
    } else if (colorBy === 'priceToSalesNtm') {
        minColorVal = 0
        maxColorVal = 30
        minColor = 'green'
        maxColor = 'red'
    } else {
        minColorVal = -65
        maxColorVal = 100
    }
    // Define the options for the portfolio map chart.
    var mapOptions = {
        minColor: minColor,
        midColor: midColor,
        maxColor: maxColor,
        minColorValue: minColorVal,
        maxColorValue: maxColorVal,
        showScale: true,
        textStyle: { color: 'black',
            fontSize: 16,
            bold: true
        },
        generateTooltip: showFullTooltip
    }

    // Show a customized tooltip.
    function showFullTooltip(row) {
        var sizeByLabel = PortfolioMapSizeSelectOptions.find(o => o.value === sizeBy).label
        var colorByLabel = PortfolioMapColorSelectOptions.find(o => o.value === colorBy).label
        // Format the data and labels based on what is displayed.
        var sizeRowVal = (sizeBy === 'marketValue') ? toUSD(data[row + 1][2]) : toPercent(data[row + 1][2])
        var colorRowVal = (colorBy === 'priceToSalesNtm') ? data[row + 1][3] : toPercent(data[row + 1][3])
        return '<div style="background:grey; padding:10px; border-style:solid">' +
            '<span><b>' + data[row + 1][0] + '</b></span><br>' +
            sizeByLabel + ': ' + sizeRowVal + '<br>' +
            colorByLabel + ': ' + colorRowVal + '<br>';
    }

    return (
        <div className="portfoliomap-container" align="center">
            <Chart
                chartType="TreeMap"
                width="1600px"
                height="750px"
                data={data}
                options={mapOptions}
            />
        </div>
    );
}