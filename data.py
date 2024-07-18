import matplotlib.pyplot as plt


y1 = [4020785, 7717126, 11414229, 15110824, 18807546, 22504268, 26200990, 29897712, 33594434, 37291156 ]
y2 = [472082, 793458, 1115058, 1436509, 1757997, 2079485, 2400973, 2722461, 3043949, 3365437]
x = [10*i for i in range(1,11)]
plt.plot(x, y1, marker='o', label="RegDomain")
plt.plot(x, y2, marker='x', label="RegCluster")

plt.title('Registration cost in broadcast mailing')
plt.xlabel('Number of domain/cluster members')
plt.ylabel('Gas cost')
plt.legend()
plt.grid()
plt.show()
