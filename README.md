======================================================================
OptimalCrop -- Go library for intelligent image re-sizing and cropping
======================================================================

OptimalCrop is a Go library for croping and re-sizing the image by locating the interesting parts.

Testing
---------

Once you build the project you can use imageResizer:

    imageResizer -in inputDir -out outputDir -width 128 -height 128



How It Works
--------------	
 
The interestingness of an image is subjective and may vary from one person to another. One way to quantitatively measure the interestingness is to measure the informations contained in that image. I thought that an interesting region of an image is a zone that carries a lot of informations. We need to be able to calculate the information at each individual pixel of the image to find out the information of a particular region.
One could calculate the information of the pixel based on information theoretic definition:

	I = -log(p(i))
	where i is the pixel, I is the self information and p(i) is the probability of occurrence of the pixel i in our image.
	
The probability of occurrence is simply the frequency of that particular pixel. An efficient way to calculate the probability is using a normalized histogram. The histogram stores the frequency of an intensity measure of the pixel. In our case we convert the RGB image to the CIELAB color space. A color space invented by the CIE (Commission internationale de l'éclairage). It describes all the colors visible to the human eye.
The problem is reduced to maximizing the total information in a region R(h,w). Or equivalently to find a region of width w and height h with max information. In order to find that region we compute the information per line (i.e. the sum of the info of the pixels in that line) and the information per column.
For this reason you need only to know how to find the maximum sum subsequence for the lines and the columns. Fortunately this is a well known problem that can be solved in linear time O(n).
