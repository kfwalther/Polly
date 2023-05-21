import './PortfolioMapChart.css'
import { Chart } from 'react-google-charts';
import { toPercent, toUSD } from "./Helpers";

export default function PortfolioMapChart({ chartData }) {

    function getFilteredChartData() {
        // Filter for only non-zero securities.
        let filtered = chartData.filter(s => (s.marketValue > 0.0 && s.securityType === 'Stock'));
        // Put the filtered values in an array, and add a column header.
        let data = filtered.map(s => [s.ticker, s.securityType, s.marketValue, s.revenueGrowthPercentageYoy * 100])
        data.unshift(['Stock', null, 0, 0])
        data.unshift(['Ticker', 'Type', 'Market Value', 'Revenue Growth %'])
        console.log(data)
        return data
    }

    // Filter the map data.
    var data = getFilteredChartData()
    // Define the options for the portfolio map chart.
    var mapOptions = {
        minColor: 'red',
        midColor: 'grey',
        maxColor: 'green',
        minColorValue: -100,
        maxColorValue: 100,
        showScale: true,
        generateTooltip: showFullTooltip
    }

    // Show a customized tooltip.
    function showFullTooltip(row) {
        return '<div style="background:grey; padding:10px; border-style:solid">' +
            '<span><b>' + data[row + 1][0] + '</b></span><br>' +
            'Market Value: ' + toUSD(data[row + 1][2]) + '<br>' +
            'Revenue Growth: ' + toPercent(data[row + 1][3]) + '<br>';
    }

    return (
        <div className="portfoliomap-container" align="center">
            <Chart
                chartType="TreeMap"
                width="1000px"
                height="650px"
                data={data}
                options={mapOptions}
            />
        </div>
    );
}