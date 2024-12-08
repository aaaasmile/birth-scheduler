#!/bin/bash
read -p "Update the birthday-scheduler (birthday-scheduler on invido.it) y/n ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]
then
	echo "Update canceled."
	exit 0

fi
echo "Stop the service"
sudo systemctl stop birthday-scheduler

ZIPDIR="./zips"
CURRDIR="./current"
OLDDIR="./old"

echo "Now starting the process..."

# Make sure dir exits
[ ! -d "$ZIPDIR" ] && mkdir -p "$ZIPDIR"
[ ! -d "$CURRDIR" ] && mkdir -p "$CURRDIR" 
[ ! -d "$OLDDIR" ] && mkdir -p "$OLDDIR" 

# backup the current dir
bckdir=$OLDDIR'/'"$(date +"%Y-%m-%d-%H%M%S")"
echo "Backup dir is: $bckdir"
[ ! -d "$bckdir" ] && mkdir -p "$bckdir" 

mv $CURRDIR'/templates' $bckdir
mv $CURRDIR'/'*.bin $bckdir
mv $CURRDIR'/config.toml' $bckdir

#zips=$(ls $ZIPDIR)
#echo "$zips"

for file in $ZIPDIR/*.zip ; do 
	fname=$(basename "$file")
	#echo "Name is $fname"
done

echo "Want to unzip $fname"
zippath=$ZIPDIR'/'$fname
destpath=$CURRDIR'/'
echo "The source is $zippath and destination is $destpath"
unzip $zippath -d $destpath

chmod +x $destpath'/'birthday-scheduler.bin

echo "Start the service"
sudo systemctl start birthday-scheduler


echo "That's all folks!"
