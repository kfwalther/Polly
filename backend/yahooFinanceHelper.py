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

# Get the input symbol.
symbol = sys.argv[1].strip()
# Check if a symbol was provided, then query Yahoo finance.
if symbol and not symbol.isspace():
    stock = yf.Ticker(symbol)
    print(json.dumps(stock.info, indent=4))
