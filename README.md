# Heliostat Project

Work in progress... 

Basic architecture:

[Robot] <-> [GRBL on Arduion] <-> [RPI] <-> Terminal Client 

Robot has two steppers, each controlling azimuth and altitude of a mirror. 

GRBL is used to send signals to the driver boards.

GRBL recieves GCODE from RPI, which hosts the ws server(and serves a control panel site)