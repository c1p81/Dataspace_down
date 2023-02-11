# Dataspace Sentinel Downloader

This software search and download Sentinel images from Dataspace with new API  
[Documentation](https://documentation.dataspace.copernicus.eu/#/APIs/OData)  
Tested on Linux Debian testing with go 1.20  
Requires **curl**


 
By default, the software only searches. To activate the download set the switch to true  
Help : **dataspace_down.go -h**  
Example : **go run dataspace_down.go --username=my_username --password=my_password --download=true**  
# Parameters  
  **-cloudCover** string  
    	Less than % cloud cover (default "10")  
  **-collection** string  
    	Collection (default "SENTINEL-2")  
  **-dest_path** string  
    	Download folder (default "./")  
  **-download**  
    	If true start the download  
  **-end_date** string  
    	End sensing date (default Now) Format YYYY-MM-DDThh:mm:ss.000Z  
  **-password** string  
    	Password (required)  
  **-search_point**_lat string  
    	Latitude (default "43.78186592737776")  
  **-search_point_lon** string  
    	Longitude (default "11.287615415088597")  
  **-start_date** string   
    	Start sensing date (default Now - 5 days) Format YYYY-MM-DDThh:mm:ss.000Z  
  **-username** string  
    	Username (required)  

