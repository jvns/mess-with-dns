import dash
from dash import dcc, html, dash_table
import dash_bootstrap_components as dbc
import plotly.express as px
import pandas as pd
import json
import os

app = dash.Dash(__name__)

pg_connection_string = os.environ['PG_CONNECTION_STRING']

def subdomains_graph():
    df = pd.read_sql_query('SELECT * FROM subdomains', pg_connection_string)
    df['created_at'] = pd.to_datetime(df.created_at)
    df['created_at'] = df['created_at'].dt.round('D')
    df_by_month = df.groupby(['created_at']).count()
    return px.line(df_by_month.reset_index(), x='created_at', y='name')

def popular_records_graph():
    df = pd.read_sql_query('SELECT subdomain, count(*) AS count FROM dns_records group by 1 order by 2 desc limit 30', pg_connection_string)
    # make name column a link
    df['subdomain'] = df['subdomain'].apply(lambda x: html.A(x, href=f'/{x}'))
    return dbc.Table.from_dataframe(df, striped=True, bordered=True, hover=True)

def popular_requests_graph():
    df = pd.read_sql_query('SELECT subdomain, count(*) AS count FROM dns_requests group by 1 order by 2 desc limit 30', pg_connection_string)
    df['subdomain'] = df['subdomain'].apply(lambda x: html.A(x, href=f'/{x}'))
    return dbc.Table.from_dataframe(df, striped=True, bordered=True, hover=True)

def remove_hdr(x):
    x.pop('Hdr')
    return x

def parse_response(x):
    answers = json.loads(x)['Answer']
    if answers is None or len(answers) == 0:
        return 'no answer'
    return str(remove_hdr(answers[0]))

def get_dns_requests_table(subdomain):
    df = pd.read_sql_query(f"SELECT name, src_host, response FROM dns_requests where subdomain = '{subdomain}' LIMIT 100", pg_connection_string)
    df['response'] = df['response'].apply(parse_response)
    return dash_table.DataTable(
        id='table',
        columns=[{"name": i, "id": i} for i in df.columns],
        data=df.to_dict('records'),
    )

def get_dns_records_table(subdomain):
    df = pd.read_sql_query(f"SELECT name, content FROM dns_records where subdomain = '{subdomain}' LIMIT 50", pg_connection_string)
    df['content'] = df['content'].apply(lambda x: str(remove_hdr(json.loads(x))))
    return dash_table.DataTable(
        id='table',
        columns=[{"name": i, "id": i} for i in df.columns],
        data=df.to_dict('records'),
    )


homepage_layout = html.Div(children=[
    html.H1(children='mess with dns'),
    dcc.Location(id='url', refresh=False),
    html.Div(children='''
    some metrics lol
    '''),

    dcc.Graph(
        id='subdomains',
        figure=subdomains_graph()
    ),

   html.H2('popular records'),
   popular_records_graph(),
   html.H2('popular requests'),
   popular_requests_graph(),
])

app.layout = html.Div([
    dcc.Location(id='url', refresh=False),
    html.Div(id='page-content')
])

@app.callback(dash.dependencies.Output('page-content', 'children'),
              [dash.dependencies.Input('url', 'pathname')])
def display_page(pathname):
    if pathname == '/':
        return homepage_layout
    return html.Div([
        html.H1('{}'.format(pathname)),
        html.H2('dns records'),
        get_dns_records_table(pathname.split('/')[-1]),
        html.H2('dns requests'),
        get_dns_requests_table(pathname.split('/')[-1]),
    ])
# regenerate graph when the page is reloaded


if __name__ == '__main__':
    app.run_server(debug=True)
