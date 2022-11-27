#version 330

in vec3 in_position;

in vec2 pos1;
in float lineLength;
in vec2 sinCos;

uniform mat4 projection;
uniform float scale;
uniform mat4 distort;

mat2 rotation2d(vec2 sinCos) {
    return mat2(
        sinCos.y, -sinCos.x,
        sinCos.x, sinCos.y
    );
}

void main() {
    vec2 posB = rotation2d(sinCos) * vec2(in_position.x*lineLength, in_position.y*scale);

    gl_Position = projection * distort * vec4(posB+pos1, in_position.z, 1.0);
}