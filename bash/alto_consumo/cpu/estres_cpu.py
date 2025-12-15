def numero_primo(n):
    if n <= 1:
        return False
    
    for i in range(2, int(n**0.5) + 1):
        if n % i == 0:
            return False
    return True


if __name__ == '__main__':
    i = 2
    while True:
        resultado = numero_primo(i)
        print(f'El numero: {i} es un numero primo? {resultado}')
        i += 1