'''
Python script that uses yfinance to extract the encrypted data stores from the Yahoo finance web page. It does this using 
the BeautifulSoup web scraper.

Tested with Python:
3.11.3
pip install yfinance --upgrade --no-cache-dir
'''

import json
import sys
import yfinance as yf

# Get the input symbols.
symbols = sys.argv[1].strip()
# Check if a symbol was provided, then query Yahoo finance.
if symbols and not symbols.isspace():
    # Put spaces b/w symbols instead of commas.
    symbolsSpaced = ' '.join(symbols.split(','))
    # Get the info for all symbols.
    stockInfoDict = yf.Tickers(symbolsSpaced)
    # Use dict-comprehension to pull out the 'info' portion of the Ticker object into new dict.
    reducedStockInfoDict = {ticker: stockInfo.info for ticker, stockInfo in stockInfoDict.tickers.items()}
    # Iterate thru the list of Ticker objects, and dump their JSON-formatted info.
    print(json.dumps(reducedStockInfoDict, indent=4))
