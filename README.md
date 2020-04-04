This is a simple application that pulls the ServiceNow API for any open and unassiged ticket for a specific queue. It in tern sends a webhook reqest to a Luxafor LED Flag to indicate waiting tickets in queue.

Defaut LED colore are (but can be overridded with the low and hight value areguments:  
0 Open or Unassiged tickets = OFF  
1 Open or Unassiged tickets = Green  
2 Open or Unassiged tickets = Blue  
3 or more Open or Unassiged tickets = Red  

Running the application requiues the following arguments or enviroment variables: 

Command like arguments: 
- -l (Luxafor API ID)
- -u (ServiceNow Username)
- -p (ServiceNow Password)
- -a (ServiceNow Assignment Group)

Enviroment variables: 
- SNF_LUXID (Luxafor API I)
- SNF_SNUSER (ServiceNow Username)
- SNF_SNPASS (ServiceNow Passwor)
- SNF_SNASSIGNGROUP (ServiceNow Assignment Group)


Optionnal arguments:
- -c / SNF_SNBASEURL (Custom base ServiceNow url)
- -v / SNF_VERBOSE (Verbose mode)
- -low / SNF_LOW (Low value for number of serivce now tickets)
- -high / SNF_HIGH (High value for number of serivce now tickets)

Updated 2020-04-04  
